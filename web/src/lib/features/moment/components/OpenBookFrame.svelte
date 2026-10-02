<svg
	class="open-book-frame"
	viewBox="0 0 1600 900"
	preserveAspectRatio="none"
	aria-hidden="true"
	focusable="false"
>
	<defs>
		<linearGradient id="moment-cover" x1="0" y1="0" x2="0" y2="1">
			<stop offset="0" stop-color="#9a6542" /><stop offset="0.5" stop-color="#70452f" /><stop
				offset="1"
				stop-color="#44281d"
			/>
		</linearGradient>
		<linearGradient id="moment-cover-light" x1="0" y1="0" x2="1" y2="0">
			<stop offset="0" stop-color="#f0c58e" stop-opacity="0.3" /><stop
				offset="0.18"
				stop-color="#fff0cb"
				stop-opacity="0.06"
			/><stop offset="0.78" stop-color="#1f100b" stop-opacity="0.02" /><stop
				offset="1"
				stop-color="#1f100b"
				stop-opacity="0.28"
			/>
		</linearGradient>
		<linearGradient id="moment-left-page" x1="0" y1="0" x2="1" y2="0">
			<stop offset="0" stop-color="#e2d4bc" /><stop offset="0.055" stop-color="var(--paper)" /><stop
				offset="0.88"
				stop-color="var(--paper)"
			/><stop offset="1" stop-color="var(--paper-deep)" />
		</linearGradient>
		<linearGradient id="moment-right-page" x1="0" y1="0" x2="1" y2="0">
			<stop offset="0" stop-color="var(--paper-deep)" /><stop
				offset="0.12"
				stop-color="var(--paper)"
			/><stop offset="0.94" stop-color="var(--paper)" /><stop offset="1" stop-color="#e2d4bc" />
		</linearGradient>
		<linearGradient id="moment-spine" x1="0" y1="0" x2="1" y2="0">
			<stop offset="0" stop-color="#58402d" stop-opacity="0" /><stop
				offset="0.38"
				stop-color="#5b402c"
				stop-opacity="0.5"
			/><stop offset="0.55" stop-color="#fff3d4" stop-opacity="0.35" /><stop
				offset="1"
				stop-color="#66472f"
				stop-opacity="0"
			/>
		</linearGradient>
		<filter id="moment-book-shadow" x="-15%" y="-18%" width="135%" height="150%"
			><feDropShadow
				dx="0"
				dy="24"
				stdDeviation="23"
				flood-color="#170d08"
				flood-opacity="0.62"
			/></filter
		>
		<filter id="moment-page-lift" x="-8%" y="-8%" width="116%" height="120%"
			><feDropShadow
				dx="0"
				dy="5"
				stdDeviation="6"
				flood-color="#3f2819"
				flood-opacity="0.2"
			/></filter
		>
	</defs>

	<ellipse cx="800" cy="848" rx="750" ry="54" fill="#160c08" opacity="0.4" />
	<g filter="url(#moment-book-shadow)">
		<path
			d="M25 54Q260 5 795 43Q1120 7 1573 51L1592 865Q1190 900 800 880Q410 900 7 860Z"
			fill="url(#moment-cover)"
			stroke="#321c14"
			stroke-width="5"
			vector-effect="non-scaling-stroke"
		/>
		<path
			d="M34 59Q260 18 794 54Q1120 19 1563 60L1580 851Q1182 884 800 866Q415 884 18 849Z"
			fill="url(#moment-cover-light)"
			stroke="#b27a50"
			stroke-opacity="0.35"
			stroke-width="1.3"
			vector-effect="non-scaling-stroke"
		/>
		<path
			d="M42 68Q264 32 790 63L797 854Q418 856 35 838Z"
			fill="#cbb58d"
			stroke="#9e805b"
			stroke-width="1.2"
			vector-effect="non-scaling-stroke"
		/>
		<path
			d="M810 62Q1120 28 1555 68L1571 839Q1185 870 803 854Z"
			fill="#cbb58d"
			stroke="#9e805b"
			stroke-width="1.2"
			vector-effect="non-scaling-stroke"
		/>
		<g filter="url(#moment-page-lift)">
			<path
				d="M38 47Q275 8 789 51Q798 448 800 850Q420 817 31 837Q19 445 38 47Z"
				fill="url(#moment-left-page)"
				stroke="#c7af85"
				stroke-width="1.5"
				vector-effect="non-scaling-stroke"
			/>
			<path
				d="M811 50Q1122 12 1560 48Q1578 443 1568 834Q1180 813 800 850Q806 448 811 50Z"
				fill="url(#moment-right-page)"
				stroke="#cfb78d"
				stroke-width="1.5"
				vector-effect="non-scaling-stroke"
			/>
		</g>
		<path d="M765 48Q800 21 836 52L829 852Q802 822 775 850Z" fill="url(#moment-spine)" />
		<path
			d="M791 52Q783 451 800 849M810 51Q821 451 800 849"
			fill="none"
			stroke="#72543b"
			stroke-opacity="0.48"
			stroke-width="2"
			vector-effect="non-scaling-stroke"
		/>
		<path
			d="M801 52Q800 451 800 849"
			fill="none"
			stroke="#fff4d7"
			stroke-opacity="0.42"
			stroke-width="1.2"
			vector-effect="non-scaling-stroke"
		/>
		<g fill="none" stroke="#876d4d" stroke-opacity="0.38" vector-effect="non-scaling-stroke">
			<path
				d="M47 833Q420 814 787 846M52 839Q420 821 784 851M816 851Q1184 814 1556 834M824 857Q1185 822 1548 841"
			/>
			<path d="M1559 72Q1574 447 1562 817M1565 87Q1580 450 1568 804" />
		</g>
	</g>
</svg>

<style>
	.open-book-frame {
		position: absolute;
		inset: -0.85rem -1.35rem -1.65rem -0.85rem;
		z-index: 0;
		width: calc(100% + 2.2rem);
		height: calc(100% + 2.5rem);
		overflow: visible;
		pointer-events: none;
	}
	@media (max-width: 767px) {
		.open-book-frame {
			display: none;
		}
	}
</style>
