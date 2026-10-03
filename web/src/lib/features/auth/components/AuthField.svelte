<script lang="ts">
	import Input from '$lib/ui/primitives/input/Input.svelte';
	import type { Snippet } from 'svelte';
	import type { FullAutoFill } from 'svelte/elements';
	import Eye from 'lucide-svelte/icons/eye';
	import EyeOff from 'lucide-svelte/icons/eye-off';

	let {
		label,
		name,
		type = 'text',
		autocomplete,
		required = false,
		placeholder = '',
		icon
	}: {
		label: string;
		name: string;
		type?: string;
		autocomplete?: FullAutoFill;
		required?: boolean;
		placeholder?: string;
		icon?: Snippet;
	} = $props();
	let visible = $state(false);
</script>

<div class="auth-field">
	<label for={name}>{label}</label>
	<div class="auth-field-control">
		<Input
			id={name}
			{name}
			type={type === 'password' && visible ? 'text' : type}
			{autocomplete}
			{required}
			{placeholder}
			{icon}
			inputClass={type === 'password' ? 'auth-input password-input' : 'auth-input'}
		/>
		{#if type === 'password'}
			<button
				type="button"
				class="password-toggle"
				aria-label={visible ? '隐藏密码' : '显示密码'}
				aria-pressed={visible}
				onclick={() => (visible = !visible)}
			>
				{#if visible}<EyeOff size={18} />{:else}<Eye size={18} />{/if}
			</button>
		{/if}
	</div>
</div>

<style lang="postcss">
	@reference "$routes/layout.css";
	.auth-field label {
		display: block;
		margin-bottom: 9px;
		color: #57534e;
		font-size: 13px;
	}
	.auth-field-control {
		position: relative;
	}
	.auth-field :global(.auth-input) {
		height: 46px;
		border: 1px solid #e7e5e4;
		border-radius: 6px;
		background: #fff !important;
		color: #292524 !important;
		font-size: 14px;
	}
	.auth-field :global(.auth-input::placeholder) {
		color: #a8a29e;
	}
	.auth-field :global(.auth-input:focus) {
		border-color: #f43f5e;
		box-shadow: 0 0 0 3px #fff1f2;
		outline: none;
	}
	.auth-field :global(.password-input) {
		padding-right: 44px;
	}
	.password-toggle {
		position: absolute;
		top: 1px;
		right: 1px;
		display: flex;
		align-items: center;
		justify-content: center;
		width: 44px;
		height: 44px;
		color: #78716c;
	}
	.password-toggle:focus-visible {
		outline: 2px solid #f43f5e;
		outline-offset: -4px;
	}
	@media (max-width: 640px) {
		.auth-field :global(.auth-input) {
			font-size: 16px;
		}
	}
</style>
