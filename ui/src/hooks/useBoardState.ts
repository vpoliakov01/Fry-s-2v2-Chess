import { useCallback, useReducer, useState } from 'react';
import { Move, movesEqual, PGNMove, Position } from '../common';
import {
	BestMoveResponse,
	GameEndedResponse,
	LoadGameResponse,
	Message,
	MessageType,
	playSound,
	SaveGameResponse,
	Sound,
} from '../utils';
import { gameReducer, loadInitialState } from './gameReducer';
import { useGameSocket } from './useGameSocket';

export function useBoardState() {
	const [state, dispatch] = useReducer(gameReducer, undefined, loadInitialState);
	const { board, activePlayer, allMoves, currentMove, selectedMove, availableMoves, score, pgn } = state;
	const moves = allMoves.slice(0, currentMove + 1);

	const [selectedSquare, setSelectedSquare] = useState<Position | null>(null);

	const handleMessage = useCallback((message: Message) => {
		switch (message.type) {
			case MessageType.AvailableMoves:
				dispatch({ type: 'setAvailableMoves', moves: (message.data as PGNMove[]).map(Move.fromPGN) });
				break;
			case MessageType.EngineMove:
				dispatch({ type: 'engineMove', moveData: message.data as BestMoveResponse });
				playSound(Sound.Move);
				break;
			case MessageType.SaveGameResponse:
				dispatch({ type: 'setPgn', pgn: (message.data as SaveGameResponse).pgn });
				break;
			case MessageType.LoadGameResponse: {
				const loadData = message.data as LoadGameResponse;
				dispatch({
					type: 'replayMoves',
					pastMoves: loadData.pastMoves.map(Move.fromPGN),
					currentMove: loadData.currentMove,
				});
				break;
			}
			case MessageType.GameEnded: {
				const gameEndedData = message.data as GameEndedResponse;
				console.log(`${gameEndedData.king} king has fallen! ${gameEndedData.winner} are victorious!`);
				playSound(Sound.GameEnd);
				break;
			}
			case MessageType.Processing:
				console.log('Thinking...');
				break;
			case MessageType.StoppedProcessing:
				console.log('Stopped thinking');
				break;
			default:
				console.log('unknown message', message);
				break;
		}
	}, []);

	const { sendMessage, connected } = useGameSocket(handleMessage);

	const setCurrentMove = useCallback((value: number) => {
		dispatch({ type: 'setCurrentMove', currentMove: value });
	}, []);

	const setSelectedMove = useCallback((value: number) => {
		dispatch({ type: 'setSelectedMove', selectedMove: value });
	}, []);

	const setViewMove = useCallback((value: number | null, expectedMoveCount?: number) => {
		dispatch({ type: 'setViewMove', currentMove: value, expectedMoveCount });
	}, []);

	const setPgn = useCallback((value: string) => {
		dispatch({ type: 'setPgn', pgn: value });
	}, []);

	const isValidMove = useCallback((move: Move): boolean => {
		return availableMoves.some(m => movesEqual(m, move));
	}, [availableMoves]);

	const movePiece = useCallback((move: Move, playerMove: boolean = false) => {
		if (playerMove) {
			if (!isValidMove(move)) {
				return false;
			}
			sendMessage(new Message(MessageType.PlayerMove, move.toPGN()));
			playSound(Sound.Move);
		}
		dispatch({ type: 'movePiece', move, playerMove });
		return true;
	}, [isValidMove, sendMessage]);

	// Plays continuation moves [1..upToIndex] from the current position, attaching the
	// remaining slice of the continuation to each appended move so it can be re-displayed.
	const playContinuationFromCurrent = useCallback((continuation: Move[], upToIndex: number) => {
		for (let i = 1; i <= upToIndex && i < continuation.length; i++) {
			const move = continuation[i];
			sendMessage(new Message(MessageType.PlayerMove, move.toPGN()));
			dispatch({ type: 'movePiece', move, playerMove: true, continuation: continuation.slice(i) });
		}
	}, [sendMessage]);

	return {
		state,
		activePlayer,
		allMoves,
		availableMoves,
		board,
		currentMove,
		selectedMove,
		moves,
		pgn,
		score,
		selectedSquare,
		setCurrentMove,
		setSelectedMove,
		setViewMove,
		movePiece,
		playContinuationFromCurrent,
		setPgn,
		setSelectedSquare,
		sendMessage,
		connected,
	};
}
