package rag

import (
	"regexp"
	"strings"

	domain "github.com/shawns-yao/shawn-blog/server/internal/domain/rag"
)

const retrievalPolicyVersion = "query-policy-v2"

var comparisonRoute = regexp.MustCompile(`(?i)比较|对比|区别|差异|\b(compare|comparison|differences?)\b`)
var globalRoute = regexp.MustCompile(`(?i)概览|概述|概括|总结|\b(summarize|summary|overview)\b`)
var proceduralRoute = regexp.MustCompile(`(?i)如何|怎么|步骤|流程|安装|配置|部署|\b(how to|step-by-step|procedure|install|configure|deploy)\b`)
var verificationRoute = regexp.MustCompile(`(?i)核实|核验|查证|\b(fact[- ]check|verify|verification)\b`)
var exactLookupRoute = regexp.MustCompile(`(?i)查找|查一下|在哪|哪里|位置|定义|字段|错误码|\b(find|locate|definition|defined|error code)\b`)
var exactIdentifier = regexp.MustCompile(`(?i)\b[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\b|\b[A-Za-z][A-Za-z0-9]*_[A-Za-z0-9_]+\b|\b[A-Za-z][A-Za-z0-9]*\.[A-Za-z][A-Za-z0-9_.]*\b`)
var binaryFactRoute = regexp.MustCompile(`(?i)^(is|are|was|were|does|did|has|have)\b`)
var personalFactReference = regexp.MustCompile(`(?i)\b(i|me|we|us)\b`)

func completeBinaryFactQuestion(question string) bool {
	question = strings.TrimSpace(question)
	return binaryFactRoute.MatchString(question) && (strings.HasSuffix(question, "?") || strings.HasSuffix(question, "？")) &&
		!complexQueryPattern.MatchString(question) && !referenceQueryPattern.MatchString(question) && !personalFactReference.MatchString(question)
}

// Rules cover explicit routes; ambiguous questions still use the existing model.
func ruleQueryPlan(question string) (queryPlan, bool) {
	p := queryPlan{Intent: domain.IntentKnowledgeQuery, Query: question, Queries: []string{question}, UnderstandingSource: "rule"}
	switch {
	case exactLookupRoute.MatchString(question) && exactIdentifier.MatchString(question):
		p.Strategy = "EXACT"
	case followupPattern.MatchString(question) || referenceQueryPattern.MatchString(question):
		return p, false
	case comparisonRoute.MatchString(question):
		p.Strategy, p.NeedMultiQuery = "COMPARE", true
	case len(literalSubqueries(question, "MULTI_HOP")) > 1 && proceduralRoute.MatchString(question):
		p.Strategy, p.NeedMultiQuery = "MULTI_HOP", true
	case globalRoute.MatchString(question):
		p.Strategy = "GLOBAL"
	case proceduralRoute.MatchString(question):
		p.Strategy = "PROCEDURAL"
	case verificationRoute.MatchString(question):
		p.Strategy = "FACT"
	case completeBinaryFactQuestion(question):
		p.Strategy = "FACT"
	default:
		return p, false
	}
	return p, true
}

func createRetrievalPolicy(t domain.Tuning, plan queryPlan) (domain.Tuning, domain.RetrievalPolicy) {
	effective := t
	p := domain.RetrievalPolicy{Version: retrievalPolicyVersion, Mode: "hybrid", Reason: "configured"}
	if t.AdaptiveRetrievalEnabled {
		switch plan.Strategy {
		case "EXACT":
			if t.RRFKeywordWeight > 0 {
				effective.VectorTopK, effective.RRFVectorWeight, effective.RRFKeywordWeight = 0, 0, 1
				p.Mode, p.Reason = "keyword", "exact_identifier"
			}
		case "FACT":
			// The SciFact comparison lost a relevant source at half-depth. Keep
			// recall capacity; final dynamic TopK still limits selected evidence.
			p.Reason = "single_fact_coverage"
		default:
			if plan.Strategy != "" {
				p.Reason = strings.ToLower(plan.Strategy) + "_coverage"
			}
		}
	}
	if t.RRFVectorWeight == 0 {
		effective.VectorTopK = 0
	}
	if t.RRFKeywordWeight == 0 {
		effective.KeywordTopK = 0
	}
	if t.AdaptiveRetrievalEnabled {
		available := (effective.VectorTopK + effective.KeywordTopK) * max(1, len(plan.Queries))
		effective.RerankCandidateTopK = min(t.RerankCandidateTopK, available)
	}
	if effective.VectorTopK == 0 {
		p.Mode = "keyword"
	}
	if effective.KeywordTopK == 0 {
		p.Mode = "semantic"
	}
	p.VectorTopK, p.KeywordTopK, p.RerankTopK = effective.VectorTopK, effective.KeywordTopK, effective.RerankCandidateTopK
	p.VectorWeight, p.KeywordWeight, p.ContextMaxTokens = effective.RRFVectorWeight, effective.RRFKeywordWeight, effective.ContextMaxTokens
	return effective, p
}

var explicitQuotedTerms = regexp.MustCompile("`([^`\\r\\n]+)`")
var structuredTerms = regexp.MustCompile(`[A-Za-z0-9][A-Za-z0-9_+#./:-]*`)
var technologyTerms = regexp.MustCompile(`(?i)\b(Go|Java|JavaScript|RAG|SQL|OpenAI)\b`)

func queryProtectedTerms(question string) []string {
	var result []string
	seen := map[string]bool{}
	add := func(term string) {
		term = strings.Trim(term, ".,:;/")
		key := strings.ToLower(term)
		if term != "" && !seen[key] {
			result = append(result, term)
			seen[key] = true
		}
	}
	for _, match := range explicitQuotedTerms.FindAllStringSubmatch(question, -1) {
		add(match[1])
	}
	for _, term := range technologyTerms.FindAllString(question, -1) {
		add(term)
	}
	for _, term := range structuredTerms.FindAllString(question, -1) {
		term = strings.Trim(term, ".,:;/")
		if strings.ContainsAny(term, "0123456789_+#./:-") || exactIdentifier.MatchString(term) {
			add(term)
		}
	}
	return result
}

func preservesQueryTerms(query string, terms []string) bool {
	query = strings.ToLower(query)
	for _, term := range terms {
		pattern := regexp.QuoteMeta(strings.ToLower(term))
		if regexp.MustCompile(`^[A-Za-z0-9_+#./:-]+$`).MatchString(term) {
			pattern = `(^|[^a-z0-9_])` + pattern + `($|[^a-z0-9_])`
		}
		if !regexp.MustCompile(pattern).MatchString(query) {
			return false
		}
	}
	return true
}
