<script lang="ts">
	import Button from '$lib/ui/primitives/button/Button.svelte';
	import UserCenterWindow from '$lib/features/user-center/components/UserCenterWindow.svelte';
	import type { MusicDensity } from '../types';
	let {
		loggedIn,
		density = $bindable('comfortable'),
		showCovers = $bindable(true),
		volume = $bindable(1)
	} = $props<{ loggedIn: boolean; density: MusicDensity; showCovers: boolean; volume: number }>();
</script>

<section class="music-settings" aria-label="音乐设置">
	{#if loggedIn}
		<div class="account-settings" role="region" aria-label="账号信息">
			<UserCenterWindow embedded />
		</div>
	{/if}
	<fieldset>
		<legend>歌曲列表</legend>
		<div class="setting-row">
			<span>列表密度</span>
			<div class="flex gap-2">
				<Button
					variant="secondary"
					class={density === 'comfortable' ? 'music-primary' : 'music-secondary'}
					aria-pressed={density === 'comfortable'}
					onclick={() => (density = 'comfortable')}>舒适</Button
				><Button
					variant="secondary"
					class={density === 'compact' ? 'music-primary' : 'music-secondary'}
					aria-pressed={density === 'compact'}
					onclick={() => (density = 'compact')}>紧凑</Button
				>
			</div>
		</div>
		<label class="setting-row"
			><span>显示歌曲封面</span><input type="checkbox" bind:checked={showCovers} /></label
		>
	</fieldset>
	<fieldset>
		<legend>播放</legend><label class="setting-row"
			><span>音量</span><span class="volume-setting"
				><input type="range" min="0" max="1" step="0.01" bind:value={volume} /><output
					class="font-mono text-xs">{Math.round(volume * 100)}%</output
				></span
			></label
		>
	</fieldset>
	<Button
		variant="secondary"
		class="music-secondary mt-6"
		onclick={() => {
			density = 'comfortable';
			showCovers = true;
			volume = 1;
		}}>恢复默认设置</Button
	>
</section>

<style>
	.music-settings {
		max-width: 680px;
	}
	.account-settings {
		padding-bottom: 28px;
		margin-bottom: 28px;
		border-bottom: 1px solid var(--music-border);
	}
	fieldset {
		border: 0;
		padding: 0;
		margin-bottom: 28px;
	}
	legend {
		width: 100%;
		padding-bottom: 14px;
		font-size: 15px;
		font-weight: 500;
	}
	.setting-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 18px;
		min-height: 72px;
		border-top: 1px solid var(--music-border);
		font-size: 13px;
	}
	input[type='checkbox'] {
		width: 18px;
		height: 18px;
	}
	.volume-setting {
		display: flex;
		align-items: center;
		gap: 14px;
		width: min(240px, 55%);
	}
	.volume-setting input {
		width: 100%;
		min-width: 0;
	}
	output {
		width: 40px;
		flex-shrink: 0;
		color: var(--music-muted);
	}
</style>
