import { request } from './http'

export interface ColumnItem {
  parentId: number | null
  id: number
  name: string
  shortUrl: string
  createdAt: string
  updatedAt: string
}

export interface TagItem {
  id: number
  name: string
  createdAt: string
  updatedAt: string
}

export interface TaxonomyNamePayload {
  name: string
}

export interface TaxonomySlugPayload {
  parentId?: number | null
  name: string
  shortUrl: string
}

export function listColumns() {
  return request<ColumnItem[]>('/columns', {
    method: 'GET',
  })
}

export function listTags() {
  return request<TagItem[]>('/tags', {
    method: 'GET',
  })
}

export function createColumn(payload: TaxonomySlugPayload) {
  return request<ColumnItem>('/admin/columns', {
    method: 'POST',
    body: payload,
  })
}

export function updateColumn(id: number, payload: TaxonomySlugPayload) {
  return request<ColumnItem>(`/admin/columns/${id}`, {
    method: 'PUT',
    body: payload,
  })
}

export function deleteColumn(id: number) {
  return request<void>(`/admin/columns/${id}`, {
    method: 'DELETE',
  })
}

export function createTag(name: string) {
  return request<TagItem>('/admin/tags', {
    method: 'POST',
    body: { name },
  })
}

export function updateTag(id: number, payload: TaxonomyNamePayload) {
  return request<TagItem>(`/admin/tags/${id}`, {
    method: 'PUT',
    body: payload,
  })
}

export function deleteTag(id: number) {
  return request<void>(`/admin/tags/${id}`, {
    method: 'DELETE',
  })
}

export interface TagContentItem {
  id: number
  title: string
  shortUrl: string
  summary: string
  createdAt: string
  updatedAt: string
}

export interface TagRelatedContents {
  moments: TagContentItem[]
}

export function getTagContents(id: number) {
  return request<TagRelatedContents>(`/tags/${id}/contents`, {
    method: 'GET',
  })
}
