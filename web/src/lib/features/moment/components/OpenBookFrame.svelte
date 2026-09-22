<svg
	class="open-book-frame"
	viewBox="0 0 1500 1000"
	preserveAspectRatio="none"
	aria-hidden="true"
	focusable="false"
>
	<defs>
		<linearGradient id="moment-cover" x1="0" y1="0" x2="0" y2="1">
			<stop offset="0" stop-color="#8b5a3b" />
			<stop offset="0.58" stop-color="#633c2a" />
			<stop offset="1" stop-color="#3f251b" />
		</linearGradient>
		<linearGradient id="moment-left-page" x1="0" y1="0" x2="1" y2="0">
			<stop offset="0" stop-color="#d8c59e" />
			<stop offset="0.08" stop-color="#e5d5b5" />
			<stop offset="0.82" stop-color="#ded0b3" />
			<stop offset="1" stop-color="#bca783" />
		</linearGradient>
		<linearGradient id="moment-right-page" x1="0" y1="0" x2="1" y2="0">
			<stop offset="0" stop-color="#c8b590" />
			<stop offset="0.045" stop-color="#e6d8bb" />
			<stop offset="0.88" stop-color="#eadfc6" />
			<stop offset="1" stop-color="#c7af84" />
		</linearGradient>
		<linearGradient id="moment-spine" x1="0" y1="0" x2="1" y2="0">
			<stop offset="0" stop-color="#8e7657" stop-opacity="0" />
			<stop offset="0.45" stop-color="#6d533b" stop-opacity="0.48" />
			<stop offset="0.55" stop-color="#fff5db" stop-opacity="0.35" />
			<stop offset="1" stop-color="#8e7657" stop-opacity="0" />
		</linearGradient>
		<pattern id="moment-page-lines" width="7" height="7" patternUnits="userSpaceOnUse">
			<path d="M0 1H7" stroke="#866f50" stroke-opacity="0.2" stroke-width="0.65" />
		</pattern>
		<filter id="moment-book-shadow" x="-12%" y="-12%" width="130%" height="135%">
			<feDropShadow dx="0" dy="22" stdDeviation="20" flood-color="#160d08" flood-opacity="0.55" />
		</filter>
	</defs>

	<g filter="url(#moment-book-shadow)">
		<path
			d="M20 34 Q155 2 296 26 Q900 4 1478 38 L1497 970 Q900 992 300 990 Q158 997 3 969 Z"
			fill="url(#moment-cover)"
			stroke="#3a2118"
			stroke-width="4"
			vector-effect="non-scaling-stroke"
		/>

		<path
			d="M304 31 Q902 12 1475 45 L1490 956 Q904 976 305 978 Z"
			fill="#d1bd94"
			stroke="#9d815a"
			stroke-width="1.5"
			vector-effect="non-scaling-stroke"
		/>
		<path
			d="M307 44 Q902 25 1471 56 L1486 948 Q902 965 308 966 Z"
			fill="url(#moment-page-lines)"
			opacity="0.78"
		/>

		<path
			d="M18 22 Q156 -1 294 31 L300 970 Q158 950 18 963 Q8 495 18 22 Z"
			fill="url(#moment-left-page)"
			stroke="#c5af89"
			stroke-width="1.3"
			vector-effect="non-scaling-stroke"
		/>
		<path
			d="M304 31 Q902 0 1478 24 Q1487 490 1477 951 Q900 939 300 970 Z"
			fill="url(#moment-right-page)"
			stroke="#cfba93"
			stroke-width="1.3"
			vector-effect="non-scaling-stroke"
		/>

		<path d="M268 25 Q300 5 332 28 L326 972 Q301 947 274 971 Z" fill="url(#moment-spine)" />
		<path
			d="M299 31 Q292 492 300 970"
			fill="none"
			stroke="#765a3f"
			stroke-opacity="0.44"
			stroke-width="2"
			vector-effect="non-scaling-stroke"
		/>
		<path
			d="M305 31 Q313 492 300 970"
			fill="none"
			stroke="#fff4d7"
			stroke-opacity="0.32"
			stroke-width="1.2"
			vector-effect="non-scaling-stroke"
		/>

		<g
			fill="none"
			stroke="#8e7451"
			stroke-opacity="0.44"
			stroke-width="0.85"
			vector-effect="non-scaling-stroke"
		>
			<path d="M1477 70 Q1483 495 1474 930" />
			<path d="M1481 82 Q1487 498 1478 919" />
			<path d="M1485 96 Q1491 500 1482 904" />
		</g>
		<g
			fill="none"
			stroke="#8e7451"
			stroke-opacity="0.28"
			stroke-width="0.75"
			vector-effect="non-scaling-stroke"
		>
			<path d="M35 962 Q160 947 286 968" />
			<path d="M55 966 Q165 953 280 972" />
			<path d="M322 970 Q902 943 1462 951" />
			<path d="M338 975 Q905 950 1446 957" />
		</g>
	</g>
</svg>

<style>
	.open-book-frame {
		position: absolute;
		inset: -0.55rem -1.05rem -1.15rem -0.65rem;
		z-index: 0;
		width: calc(100% + 1.7rem);
		height: calc(100% + 1.7rem);
		overflow: visible;
		pointer-events: none;
	}

	@media (max-width: 767px) {
		.open-book-frame {
			display: none;
		}
	}
</style>
