<script lang="ts">
	import { browser } from '$app/environment';
	import { createMutation, createQuery } from '@tanstack/svelte-query';
	import { authLogin } from '$lib/shared/actions/auth-login';
	import { preloadTurnstile } from '$lib/shared/actions/turnstile';
	import { authModalStore } from '$lib/shared/stores/authModalStore';
	import { windowStore } from '$lib/shared/stores/windowStore.svelte';
	import { websiteInfoCtx } from '$lib/features/website-info/context.js';
	import Button from '$lib/ui/primitives/button/Button.svelte';
	import {
		authorizeOAuthProvider,
		getTurnstileState,
		listOAuthProviders,
		login
	} from '$lib/features/auth/api';
	import type {
		AuthApproachState,
		LoginReq,
		LoginResp,
		OAuthProvider
	} from '$lib/features/auth/types';
	import {
		openOAuthPopup,
		saveOAuthFlowMeta,
		waitForOAuthPopupResult
	} from '$lib/features/auth/oauth-flow';
	import { setToken } from '$lib/shared/token';
	import { userStore } from '$lib/shared/stores/userStore';
	import { AuthCtx } from '$lib/features/auth/context';
	import AuthField from './AuthField.svelte';
	import AuthOAuthList from './AuthOAuthList.svelte';
	import AuthTurnstile from './AuthTurnstile.svelte';
	import UserRound from 'lucide-svelte/icons/user-round';
	import LockKeyhole from 'lucide-svelte/icons/lock-keyhole';

	const initialModel: AuthApproachState = {
		turnstile: {
			enabled: false,
			siteKey: '',
			error: ''
		},
		oauth: {
			providers: [],
			error: '',
			loadingKey: null
		},
		login: {
			loading: false,
			error: ''
		},
		showPasswordLogin: false
	};

	AuthCtx.mountModelData(() => initialModel);
	const { updateModelData } = AuthCtx.useModelActions();
	const authModel = AuthCtx.selectModelData((data) => data);

	let turnstileToken = $state('');

	const websiteName = websiteInfoCtx.selectModelData((data) => data?.website_name || 'shawn-blog');

	const hasOAuthProviders = $derived.by(() => ($authModel?.oauth.providers ?? []).length > 0);
	const showPasswordLogin = $derived.by(() => $authModel?.showPasswordLogin ?? false);
	const canSubmit = $derived.by(() => {
		const model = $authModel;
		if (!model) return true;
		const { turnstile } = model;
		return !turnstile.enabled || (!!turnstile.siteKey && !!turnstileToken);
	});

	const updateAuth = (updater: (prev: AuthApproachState) => AuthApproachState) => {
		updateModelData((prev) => updater(prev ?? initialModel));
	};

	const setLoginState = (partial: Partial<AuthApproachState['login']>) => {
		updateAuth((prev) => ({ ...prev, login: { ...prev.login, ...partial } }));
	};

	const setOAuthState = (partial: Partial<AuthApproachState['oauth']>) => {
		updateAuth((prev) => ({ ...prev, oauth: { ...prev.oauth, ...partial } }));
	};

	const setTurnstileState = (partial: Partial<AuthApproachState['turnstile']>) => {
		updateAuth((prev) => ({ ...prev, turnstile: { ...prev.turnstile, ...partial } }));
	};

	const setShowPasswordLogin = (value: boolean) => {
		console.log('setShowPasswordLogin', value);
		updateAuth((prev) => ({ ...prev, showPasswordLogin: value }));
	};

	const resetEphemeral = () => {
		updateAuth((prev) => ({
			...prev,
			login: { loading: false, error: '' },
			oauth: { ...prev.oauth, error: '', loadingKey: null },
			turnstile: { ...prev.turnstile, error: '' }
		}));
		turnstileToken = '';
	};

	const loginMutation = createMutation<LoginResp, Error, LoginReq>(() => ({
		mutationFn: async (payload) => login(payload)
	}));

	const oauthQuery = createQuery(() => ({
		queryKey: ['auth', 'oauth-providers'],
		enabled: $authModalStore.open,
		queryFn: () => listOAuthProviders(),
		retry: false,
		staleTime: Infinity,
		gcTime: Infinity,
		refetchOnWindowFocus: false
	}));

	const turnstileQuery = createQuery(() => ({
		queryKey: ['auth', 'turnstile'],
		enabled: $authModalStore.open,
		queryFn: () => getTurnstileState(),
		retry: false,
		staleTime: Infinity,
		gcTime: Infinity,
		refetchOnWindowFocus: false
	}));

	const executeLogin = async (payload: LoginReq) => loginMutation.mutateAsync(payload);

	const startOAuthLogin = async (provider: OAuthProvider) => {
		if (!browser) return;
		setOAuthState({ loadingKey: provider.key, error: '' });
		try {
			const redirectUri = `${window.location.origin}/auth/providers/${provider.key}/callback/`;
			const res = await authorizeOAuthProvider(provider.key, redirectUri);
			saveOAuthFlowMeta(res.state, {
				mode: 'login',
				provider: provider.key,
				returnTo: window.location.href,
				createdAt: Date.now()
			});
			const popup = openOAuthPopup(res.authUrl, provider.key);
			if (!popup) {
				window.location.href = res.authUrl;
				return;
			}
			const result = await waitForOAuthPopupResult({
				provider: provider.key,
				mode: 'login',
				popup
			});
			if (!result.success || !result.token || !result.user) {
				setOAuthState({
					loadingKey: null,
					error: result.error || 'OAuth 登录失败'
				});
				return;
			}
			setToken(result.token);
			userStore.setUser(result.user);
			authModalStore.close();
			setOAuthState({ loadingKey: null, error: '' });
		} catch (err) {
			setOAuthState({
				loadingKey: null,
				error: err instanceof Error ? err.message : '获取 OAuth 授权地址失败'
			});
		}
	};

	const handleTurnstileToken = (value: string) => {
		turnstileToken = value;
		setTurnstileState({ error: '' });
	};

	const handleTurnstileExpired = () => {
		turnstileToken = '';
	};

	const handleTurnstileError = () => {
		turnstileToken = '';
		setTurnstileState({ error: '人机验证失败，请重试' });
	};

	$effect(() => {
		if (!$authModalStore.open) {
			updateModelData(() => initialModel);
			return;
		}
		resetEphemeral();
	});

	$effect(() => {
		if (!$authModalStore.open) return;
		if (oauthQuery.isError) {
			const message =
				oauthQuery.error instanceof Error ? oauthQuery.error.message : '获取 OAuth 登录方式失败';
			updateAuth((prev) => ({
				...prev,
				oauth: { ...prev.oauth, providers: [], error: message },
				showPasswordLogin: true
			}));
			return;
		}
		if (oauthQuery.data) {
			const providers = Array.isArray(oauthQuery.data) ? oauthQuery.data : [];
			updateAuth((prev) => ({
				...prev,
				oauth: { ...prev.oauth, providers, error: '' },
				showPasswordLogin: providers.length === 0
			}));
		}
	});

	$effect(() => {
		if (!$authModalStore.open) return;
		if (turnstileQuery.isError) {
			setTurnstileState({ enabled: false, siteKey: '', error: '' });
			turnstileToken = '';
			return;
		}
		if (turnstileQuery.data) {
			const enabled = !!turnstileQuery.data.enabled;
			const siteKey = turnstileQuery.data.siteKey ?? '';
			updateAuth((prev) => ({
				...prev,
				turnstile: {
					...prev.turnstile,
					enabled,
					siteKey,
					error: enabled ? prev.turnstile.error : ''
				}
			}));
			if (!enabled) {
				turnstileToken = '';
			}
			if (enabled && siteKey) {
				preloadTurnstile().catch(() => {
					setTurnstileState({ error: '人机验证加载失败，请检查网络或拦截设置' });
				});
			}
		}
	});

	// Sync: when FloatingWindow is closed externally (e.g. user clicks outside),
	// reset authModalStore without re-triggering windowStore.close().
	$effect(() => {
		if (!windowStore.isOpen && $authModalStore.open) {
			authModalStore._reset();
		}
	});
</script>

<div class="auth-card">
	<h2>登录</h2>
	<p class="auth-subtitle">SIGN IN</p>
	<p class="auth-site">{$websiteName}</p>

	<div class="mt-6 space-y-5">
		<AuthOAuthList
			onSelect={startOAuthLogin}
			onToggleLogin={() => setShowPasswordLogin(!showPasswordLogin)}
		/>

		{#if showPasswordLogin || !hasOAuthProviders}
			<form
				class="space-y-5"
				use:authLogin={{
					execute: executeLogin,
					getPayload: (formEl) => {
						const data = new FormData(formEl);
						return {
							credential: String(data.get('credential') ?? ''),
							password: String(data.get('password') ?? ''),
							turnstileToken: turnstileToken || undefined
						};
					},
					onStart: () => {
						setLoginState({ loading: true, error: '' });
						setTurnstileState({ error: '' });
					},
					onSuccess: () => {
						setLoginState({ loading: false });
						if (loginMutation.data?.user) {
							userStore.setUser(loginMutation.data.user);
						}
						authModalStore.close();
					},
					onError: (err) => {
						setLoginState({
							loading: false,
							error: err instanceof Error ? err.message : '登录失败，请稍后重试'
						});
						if ($authModel?.turnstile.enabled) {
							turnstileToken = '';
							setTurnstileState({ error: '人机校验未通过，请重试' });
						}
					},
					onFinally: () => {
						setLoginState({ loading: false });
					}
				}}
			>
				<AuthField
					label="账号"
					name="credential"
					placeholder="请输入账号或邮箱"
					autocomplete="username"
					required
				>
					{#snippet icon()}<UserRound size={20} />{/snippet}
				</AuthField>

				<AuthField
					label="密码"
					name="password"
					type="password"
					autocomplete="current-password"
					required
					placeholder="请输入密码"
				>
					{#snippet icon()}<LockKeyhole size={20} />{/snippet}
				</AuthField>

				<AuthTurnstile
					onToken={handleTurnstileToken}
					onExpired={handleTurnstileExpired}
					onError={handleTurnstileError}
				/>

				<input type="hidden" name="turnstileToken" value={turnstileToken} />

				{#if $authModel?.login.error}
					<p class="auth-error" role="alert">
						{$authModel.login.error}
					</p>
				{/if}

				<Button
					class="auth-submit w-full"
					type="submit"
					loading={$authModel?.login.loading ?? false}
					disabled={!canSubmit}
				>
					{$authModel?.login.loading ? '登录中…' : '登录'}
				</Button>
			</form>
		{/if}
	</div>
</div>

<style lang="postcss">
	@reference "$routes/layout.css";
	.auth-card {
		color: #292524;
		color-scheme: light;
		@apply font-sans;
	}
	h2 {
		font-size: 26px;
		font-weight: 600;
		letter-spacing: 0.08em;
	}
	.auth-subtitle {
		margin-top: 4px;
		font-size: 12px;
		letter-spacing: 0.16em;
		color: #a8a29e;
	}
	.auth-site {
		margin-top: 12px;
		font-size: 13px;
		color: #78716c;
	}
	.auth-error {
		font-size: 13px;
		color: #be123c;
	}
	.auth-card :global(.auth-submit) {
		min-height: 46px;
		background: #f43f5e !important;
		color: #fff !important;
		border-radius: 6px;
	}
	.auth-card :global(.auth-submit:hover) {
		background: #e11d48 !important;
	}
	.auth-card :global(button:focus-visible) {
		outline: 2px solid #f43f5e;
		outline-offset: 3px;
	}
	.auth-card :global(svg) {
		flex-shrink: 0;
	}
</style>
