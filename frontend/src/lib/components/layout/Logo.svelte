<script lang="ts">
	interface Props {
		size?: number;
		class?: string;
	}

	let { size = 40, class: className = '' }: Props = $props();
</script>

<svg
	viewBox="0 0 128 128"
	xmlns="http://www.w3.org/2000/svg"
	width={size}
	height={size}
	class={className}
>
	<defs>
		<linearGradient id="logo-gradient" x1="0%" y1="0%" x2="100%" y2="100%">
			<stop offset="0%" stop-color="#FF6B35" />
			<stop offset="100%" stop-color="#F7931E" />
		</linearGradient>
	</defs>

	<!-- Rounded square background - uses accent color -->
	<rect
		x="4"
		y="4"
		width="120"
		height="120"
		rx="28"
		class="fill-[var(--color-accent,#1db954)]"
	/>

	<!-- LP-style curved arcs with gaps -->
	<g fill="none" stroke="url(#logo-gradient)" stroke-linecap="round">
		<!-- Outer arc -->
		<path d="M 24 64 A 40 40 0 0 1 64 24" stroke-width="8" />
		<path d="M 104 64 A 40 40 0 0 1 64 104" stroke-width="8" />

		<!-- Middle arc -->
		<path d="M 38 64 A 26 26 0 0 1 64 38" stroke-width="8" />
		<path d="M 90 64 A 26 26 0 0 1 64 90" stroke-width="8" />

		<!-- Inner arc -->
		<path d="M 52 64 A 12 12 0 0 1 64 52" stroke-width="8" />
		<path d="M 76 64 A 12 12 0 0 1 64 76" stroke-width="8" />
	</g>

	<!-- Center dot - LP spindle hole -->
	<circle cx="64" cy="64" r="6" fill="url(#logo-gradient)" />
</svg>
