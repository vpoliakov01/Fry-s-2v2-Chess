import { Color, MoveInfo, Piece } from '../common';
import { ArrowProps } from '../components/Arrow';
import { GameSettings } from './ws';

export type BoardPosition = (Piece | null | undefined)[][];

export interface SavedBoardState {
	board: BoardPosition;
	activePlayer: Color;
	allMoves: MoveInfo[];
	currentMove: number;
	score: number;
	pgn: string;
	drawnArrows: ArrowProps[];
}

const showLabelsOptions = ['all', 'border', 'pieces', 'moves', 'moves+', 'none'] as const;
const onMoveHoverOptions = ['set board', 'arrow', 'highlight', 'highlight+', 'none'] as const;
const moveNotationOptions = ['PGN', 'SAN', 'FAN', 'FAN+'] as const;

export type ShowLabels = typeof showLabelsOptions[number];
export type OnMoveHover = typeof onMoveHoverOptions[number];
export type MoveNotation = typeof moveNotationOptions[number];

export interface DisplaySettingsState {
	moveNotation: MoveNotation;
	showLabels: ShowLabels;
	showContinuation: boolean;
	showEvalBar: boolean;
	showEvalBarScore: boolean;
}

export class GameStateManager {
	private static readonly STORAGE_KEY = 'chess-board-state';
	private static readonly SETTINGS_KEY = 'chess-game-settings';
	private static readonly DISPLAY_SETTINGS_KEY = 'chess-display-settings';
	private static readonly UNDEFINED_MARKER = '__UNDEFINED__';

	static readonly defaultSettings: GameSettings = {
		humanPlayers: [0, 2],
		depth: 12,
		spread: 8,
		spreadDrop: 2,
		evalLimit: 0,
	};

	static readonly defaultDisplaySettings: DisplaySettingsState = {
		moveNotation: 'FAN+',
		showLabels: 'pieces',
		showContinuation: true,
		showEvalBar: true,
		showEvalBarScore: false,
	};

	static load(): SavedBoardState {
		try {
			const stored = localStorage.getItem(GameStateManager.STORAGE_KEY);
			if (!stored) {
				return GameStateManager.getDefault();
			}
			const parsed = JSON.parse(stored);
			return GameStateManager.deserialize(parsed);
		} catch {
			return GameStateManager.getDefault();
		}
	}

	static save(state: SavedBoardState): void {
		try {
			const serialized = GameStateManager.serialize(state);
			localStorage.setItem(GameStateManager.STORAGE_KEY, JSON.stringify(serialized));
		} catch (error) {
			console.warn('Failed to save state:', error);
		}
	}

	static clear(): void {
		localStorage.removeItem(GameStateManager.STORAGE_KEY);
	}

	static loadSettings(): GameSettings {
		try {
			const stored = localStorage.getItem(GameStateManager.SETTINGS_KEY);
			if (!stored) return GameStateManager.defaultSettings;
			return { ...GameStateManager.defaultSettings, ...JSON.parse(stored) };
		} catch {
			return GameStateManager.defaultSettings;
		}
	}

	static saveSettings(settings: GameSettings): void {
		try {
			localStorage.setItem(GameStateManager.SETTINGS_KEY, JSON.stringify(settings));
		} catch (error) {
			console.warn('Failed to save settings:', error);
		}
	}

	static loadDisplaySettings(): DisplaySettingsState {
		try {
			const stored = localStorage.getItem(GameStateManager.DISPLAY_SETTINGS_KEY);
			if (!stored) return GameStateManager.defaultDisplaySettings;
			return { ...GameStateManager.defaultDisplaySettings, ...JSON.parse(stored) };
		} catch {
			return GameStateManager.defaultDisplaySettings;
		}
	}

	static saveDisplaySettings(settings: DisplaySettingsState): void {
		try {
			localStorage.setItem(GameStateManager.DISPLAY_SETTINGS_KEY, JSON.stringify(settings));
		} catch (error) {
			console.warn('Failed to save display settings:', error);
		}
	}

	private static getDefault(): SavedBoardState {
		return {
			board: Array(14).fill(null).map(() => Array(14).fill(undefined)),
			activePlayer: Color.Red,
			allMoves: [],
			currentMove: 0,
			score: 0,
			pgn: '',
			drawnArrows: [],
		};
	}

	private static serialize(state: SavedBoardState): any {
		return {
			...state,
			board: state.board.map(row => row.map(cell => cell === undefined ? GameStateManager.UNDEFINED_MARKER : cell)),
		};
	}

	private static deserialize(data: any): SavedBoardState {
		const deserializeContinuation = (raw: any): MoveInfo[] =>
			Array.isArray(raw)
				? raw.map((m: any) => new MoveInfo(m.from, m.to, m.piece, m.capturedPiece, m.score ?? null))
				: [];

		return {
			board: Array.isArray(data.board)
				? data.board.map((row: any) =>
					Array.isArray(row)
						? row.map((cell: any) => cell === GameStateManager.UNDEFINED_MARKER ? undefined : cell)
						: Array(14).fill(undefined)
				)
				: GameStateManager.getDefault().board,
			activePlayer: data.activePlayer || Color.Red,
			allMoves: Array.isArray(data.allMoves)
				? data.allMoves.map((move: any) => {
					const info = new MoveInfo(move.from, move.to, move.piece, move.capturedPiece, move.score ?? null);
					if (Array.isArray(move.continuation)) {
						info.continuation = deserializeContinuation(move.continuation);
					}
					return info;
				})
				: [],
			currentMove: data.currentMove || 0,
			score: data.score || 0,
			pgn: data.pgn || '',
			drawnArrows: Array.isArray(data.drawnArrows) ? data.drawnArrows : [],
		};
	}
}
