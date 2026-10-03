export type MusicView = 'search' | 'playlists' | 'charts' | 'favorites' | 'settings';
export type MusicDensity = 'comfortable' | 'compact';
export type MusicPlaybackMode = 'sequence' | 'single' | 'repeat' | 'shuffle';

export interface MusicSong {
	id: string;
	title: string;
	artist: string;
	album: string;
	albumId: string;
	coverArt?: string;
	duration: number;
	track: number;
	year: number;
	public: boolean;
	unavailable?: boolean;
}

export interface MusicAlbum {
	id: string;
	name: string;
	artist: string;
	coverArt?: string;
	songCount: number;
	song?: MusicSong[];
}

export interface MusicCatalog {
	songs: MusicSong[];
	albums: MusicAlbum[];
	hasMore: boolean;
}

export interface MusicStatus {
	enabled: boolean;
	configured: boolean;
	available: boolean;
	message?: string;
	maxBitRate: number;
	maxUploadBytes: number;
}

export interface MusicLyrics {
	displayArtist?: string;
	displayTitle?: string;
	lang: string;
	synced: boolean;
	offset: number;
	line: { start?: number; value: string }[];
}

export interface MusicSongPage {
	songs: MusicSong[];
	hasMore: boolean;
}

export interface MusicPlaylist {
	id: number;
	name: string;
	songCount: number;
	updatedAt: string;
}

export interface MusicTrackProps {
	viewer: number;
	privateAccess: boolean;
	currentId?: string;
	density: MusicDensity;
	showCovers: boolean;
	play: (song: MusicSong, songs: MusicSong[]) => void;
	enqueue: (song: MusicSong) => void;
	login: () => void;
}

export interface MusicLyricsResult {
	lyrics: MusicLyrics[];
}
