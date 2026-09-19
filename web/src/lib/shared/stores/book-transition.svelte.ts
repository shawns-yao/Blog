export type BookTransitionPhase = 'idle' | 'picking' | 'opening' | 'revealing';

export interface BookTransitionPayload {
	title: string;
	color: string;
	edge: string;
	from: {
		top: number;
		left: number;
		width: number;
		height: number;
	};
}

class BookTransitionState {
	active = $state(false);
	reading = $state(false);
	pageTurning = $state(false);
	closing = $state(false);
	phase = $state<BookTransitionPhase>('idle');
	book = $state<BookTransitionPayload | null>(null);

	start(payload: BookTransitionPayload) {
		this.book = payload;
		this.phase = 'picking';
		this.active = true;
	}

	setPhase(phase: Exclude<BookTransitionPhase, 'idle' | 'picking'>) {
		if (!this.active) return;
		this.phase = phase;
	}

	enterReadingMode() {
		this.reading = true;
	}

	leaveReadingMode() {
		this.reading = false;
		this.pageTurning = false;
		this.closing = false;
	}

	startPageTurn() {
		if (!this.reading || this.pageTurning || this.closing) return false;
		this.pageTurning = true;
		return true;
	}

	finishPageTurn() {
		this.pageTurning = false;
	}

	startClosing() {
		if (!this.reading || this.closing) return false;
		this.pageTurning = false;
		this.closing = true;
		return true;
	}

	reset() {
		this.active = false;
		this.phase = 'idle';
		this.book = null;
	}
}

export const bookTransition = new BookTransitionState();
