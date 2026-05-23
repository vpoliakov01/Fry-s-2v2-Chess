import { BOARD_SIZE, Color, CORNER_SIZE, formatNumber, Move, MoveInfo, PieceType, PlayerColors } from '../common';
import { BestMoveResponse, BoardPosition, GameStateManager } from '../utils';

function convertContinuationToMoveInfo(continuationPGN: string[], startingBoard: BoardPosition): MoveInfo[] {
	const board = startingBoard.map(row => [...row]);
	const moveInfos: MoveInfo[] = [];

	for (const pgn of continuationPGN) {
		const move = Move.fromPGN(pgn);
		const piece = board[move.from.row][move.from.col];
		const capturedPiece = board[move.to.row][move.to.col] ?? null;

		moveInfos.push(new MoveInfo(move.from, move.to, piece!, capturedPiece));

		// Update board for next move
		board[move.to.row][move.to.col] = board[move.from.row][move.from.col];
		board[move.from.row][move.from.col] = null;
	}

	return moveInfos;
}

function replayBoard(allMoves: MoveInfo[], upToMoveIndex: number): BoardPosition {
	const board = createEmptyBoard();
	for (let i = 0; i <= upToMoveIndex && i < allMoves.length; i++) {
		const { from, to } = allMoves[i];
		board[to.row][to.col] = board[from.row][from.col];
		board[from.row][from.col] = null;
	}
	return board;
}

export function createEmptyBoard(): BoardPosition {
	const board: BoardPosition = Array(BOARD_SIZE).fill(null).map(() => Array(BOARD_SIZE).fill(null));

	for (let i = 0; i < CORNER_SIZE; i++) {
		for (let j = 0; j < CORNER_SIZE; j++) {
			for (const k of [i, BOARD_SIZE - 1 - i]) {
				for (const l of [j, BOARD_SIZE - 1 - j]) {
					board[k][l] = undefined;
				}
			}
		}
	}

	return setPieces(board);
}

function setPieces(board: BoardPosition): BoardPosition {
	const placePieces = (color: Color) => {
		const pieces: PieceType[][] = [[
			PieceType.Rook,
			PieceType.Knight,
			PieceType.Bishop,
			PieceType.Queen,
			PieceType.King,
			PieceType.Bishop,
			PieceType.Knight,
			PieceType.Rook,
		], Array(8).fill(PieceType.Pawn)];

		const transformIndex = (color: Color, i: number, j: number) => {
			switch (color) {
				case Color.Yellow:
					return [i, BOARD_SIZE - CORNER_SIZE - 1 - j];
				case Color.Green:
					return [BOARD_SIZE - CORNER_SIZE - 1 - j, BOARD_SIZE - 1 - i];
				case Color.Red:
					return [BOARD_SIZE - 1 - i, j + CORNER_SIZE];
				case Color.Blue:
					return [j + CORNER_SIZE, i];
				default:
					throw new Error(`Invalid color: ${color}`);
			}
		};

		for (let i = 0; i < pieces.length; i++) {
			for (let j = 0; j < pieces[i].length; j++) {
				const piece = pieces[i][j];
				const [k, l] = transformIndex(color, i, j);
				board[k][l] = { type: piece, color };
			}
		}
	};

	for (const color of PlayerColors) {
		placePieces(color);
	}

	return board;
}

export interface GameState {
	board: BoardPosition;
	activePlayer: Color;
	allMoves: MoveInfo[];
	currentMove: number;
	availableMoves: Move[];
	score: number;
	pgn: string;
}

export type GameAction =
	| { type: 'movePiece', move: Move, playerMove?: boolean, continuation?: Move[] }
	| { type: 'engineMove', moveData: BestMoveResponse }
	| { type: 'setAvailableMoves', moves: Move[] }
	| { type: 'setScore', score: number }
	| { type: 'setPgn', pgn: string }
	| { type: 'setCurrentMove', currentMove: number }
	| { type: 'setViewMove', currentMove: number }
	| { type: 'replayMoves', pastMoves: Move[], currentMove: number }
	| { type: 'reset' };

export function defaultGameState(): GameState {
	return {
		board: createEmptyBoard(),
		activePlayer: Color.Red,
		allMoves: [],
		currentMove: 0,
		availableMoves: [],
		score: 0,
		pgn: '',
	};
}

export function loadInitialState(): GameState {
	const saved = GameStateManager.load();
	return {
		board: saved.board,
		activePlayer: saved.activePlayer,
		allMoves: saved.allMoves,
		currentMove: saved.currentMove,
		availableMoves: [],
		score: typeof saved.score === 'number' ? saved.score : 0,
		pgn: saved.pgn,
	};
}

export function gameReducer(state: GameState, action: GameAction): GameState {
	switch (action.type) {
		case 'movePiece': {
			const { from, to } = action.move;
			const playerMove = action.playerMove ?? false;

			const baseMoves = state.allMoves.slice(0, state.currentMove + 1);
			const newMoveInfo = new MoveInfo(from, to, state.board[from.row][from.col]!, state.board[to.row][to.col] ?? null);
			if (action.continuation) {
				newMoveInfo.continuation = action.continuation;
			}
			const newAllMoves = [...baseMoves, newMoveInfo];

			if (state.currentMove < baseMoves.length - 1 && !playerMove) {
				return { ...state, allMoves: newAllMoves };
			}

			const newBoard = [...state.board.map(row => [...row])];
			newBoard[to.row][to.col] = newBoard[from.row][from.col];
			newBoard[from.row][from.col] = null;

			return {
				...state,
				board: newBoard,
				allMoves: newAllMoves,
				currentMove: state.currentMove + 1,
				activePlayer: PlayerColors[(PlayerColors.indexOf(state.activePlayer) + 1) % PlayerColors.length],
			};
		}

		case 'engineMove': {
			const { continuation, moveNumber, score, time, evaluations } = action.moveData;
			const move = continuation[0];
			const expectedFromCurrent = state.currentMove + 2;
			const expectedAtEnd = state.allMoves.length + 1;
			if (moveNumber !== expectedFromCurrent && moveNumber !== expectedAtEnd) {
				console.warn(
					`Ignoring stale engine move ${move} at moveNumber ${moveNumber} expected ${expectedFromCurrent} or ${expectedAtEnd}`,
				);
				return state;
			}
			console.log('move'.padEnd(8), 'time'.padStart(6), 'score'.padStart(9), 'evals'.padStart(8), 'avg'.padStart(6));
			console.log(
				move.padEnd(8),
				formatNumber(time, 3, 2, 's'),
				formatNumber(score, 5, 2),
				formatNumber(evaluations / 1000, 5, 2, 'k'),
				formatNumber(time / evaluations * 1e6, 4, 0, 'μs'),
			);
			const { from, to } = Move.fromPGN(continuation[0]);

			if (moveNumber === expectedFromCurrent) {
				const newMoveInfo = new MoveInfo(
					from,
					to,
					state.board[from.row][from.col]!,
					state.board[to.row][to.col] ?? null,
					score,
				);
				const newBoard = state.board.map(row => [...row]);
				newBoard[to.row][to.col] = newBoard[from.row][from.col];
				newBoard[from.row][from.col] = null;
				newMoveInfo.continuation = convertContinuationToMoveInfo(continuation, state.board);
				return {
					...state,
					board: newBoard,
					allMoves: [...state.allMoves.slice(0, state.currentMove + 1), newMoveInfo],
					currentMove: state.currentMove + 1,
					activePlayer: PlayerColors[(PlayerColors.indexOf(state.activePlayer) + 1) % PlayerColors.length],
					score,
				};
			}

			const endBoard = replayBoard(state.allMoves, state.allMoves.length - 1);
			const newMoveInfo = new MoveInfo(from, to, endBoard[from.row][from.col]!, endBoard[to.row][to.col] ?? null, score);
			newMoveInfo.continuation = convertContinuationToMoveInfo(continuation, endBoard);
			return { ...state, allMoves: [...state.allMoves, newMoveInfo], score };
		}

		case 'setAvailableMoves':
			return { ...state, availableMoves: action.moves };

		case 'setScore':
			return { ...state, score: action.score };

		case 'setPgn':
			return { ...state, pgn: action.pgn };

		case 'setCurrentMove':
			return { ...state, currentMove: action.currentMove };

		case 'setViewMove': {
			const idx = action.currentMove;
			const activePlayerIdx = ((idx + 1) % PlayerColors.length + PlayerColors.length) % PlayerColors.length;
			return {
				...state,
				board: replayBoard(state.allMoves, idx),
				currentMove: idx,
				activePlayer: PlayerColors[activePlayerIdx],
			};
		}

		case 'replayMoves': {
			const fresh = defaultGameState();
			let newBoard = createEmptyBoard();
			const newMoves: MoveInfo[] = [];
			const lastMoveIndex = Math.min(action.currentMove, action.pastMoves.length - 1);
			let boardAtIndex = newBoard;

			for (let i = 0; i < action.pastMoves.length; i++) {
				const { from, to } = action.pastMoves[i];
				const newMove = new MoveInfo(from, to, newBoard[from.row][from.col]!, newBoard[to.row][to.col] ?? null);

				const existing = state.allMoves[i];
				const matchesExisting = existing
					&& existing.from.row === from.row && existing.from.col === from.col
					&& existing.to.row === to.row && existing.to.col === to.col;
				if (matchesExisting) {
					newMove.continuation = existing.continuation;
				}

				newMoves.push(newMove);
				newBoard[to.row][to.col] = newBoard[from.row][from.col];
				newBoard[from.row][from.col] = null;

				if (i === lastMoveIndex) {
					boardAtIndex = newBoard;
					newBoard = [...newBoard.map(row => [...row])];
				}
			}

			return {
				...fresh,
				board: boardAtIndex,
				allMoves: newMoves,
				currentMove: lastMoveIndex,
				activePlayer: PlayerColors[(lastMoveIndex + 1) % PlayerColors.length],
			};
		}

		case 'reset':
			return defaultGameState();

		default:
			return state;
	}
}
