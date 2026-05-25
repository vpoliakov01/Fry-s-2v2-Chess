import React, { createContext, ReactNode, useContext, useEffect, useState } from 'react';
import { Color, Move } from '../common';
import { useArrowDrawing } from '../hooks/useArrowDrawing';
import { useBoardState } from '../hooks/useBoardState';
import { useDisplaySettings } from '../hooks/useDisplaySettings';
import { useGameSettings } from '../hooks/useGameSettings';
import { GameStateManager } from '../utils';

type HighlightedMove = { move: Move, color: Color };

type BoardStateContextType =
	& Omit<ReturnType<typeof useBoardState>, 'state'>
	& ReturnType<typeof useArrowDrawing>
	& ReturnType<typeof useDisplaySettings>
	& ReturnType<typeof useGameSettings>
	& {
		highlightedMove: HighlightedMove | null,
		setHighlightedMove: React.Dispatch<React.SetStateAction<HighlightedMove | null>>,
	};

const BoardStateContext = createContext<BoardStateContextType | null>(null);

export const useBoardStateContext = () => {
	const context = useContext(BoardStateContext);
	if (!context) {
		throw new Error('useBoardStateContext must be used within a BoardStateProvider');
	}
	return context;
};

export const BoardStateProvider = ({ children }: { children: ReactNode }) => {
	const { state, ...boardState } = useBoardState();
	const arrowDrawing = useArrowDrawing(state.activePlayer);
	const displaySettings = useDisplaySettings();
	const gameSettings = useGameSettings(boardState.sendMessage);
	const [highlightedMove, setHighlightedMove] = useState<HighlightedMove | null>(null);

	const { drawnArrows } = arrowDrawing;
	useEffect(() => {
		GameStateManager.save({
			board: state.board,
			activePlayer: state.activePlayer,
			allMoves: state.allMoves,
			currentMove: state.currentMove,
			score: state.score,
			pgn: state.pgn,
			drawnArrows,
		});
	}, [state, drawnArrows]);

	return (
		<BoardStateContext.Provider
			value={{
				...boardState,
				...arrowDrawing,
				...displaySettings,
				...gameSettings,
				highlightedMove: highlightedMove,
				setHighlightedMove,
			}}
		>
			{children}
		</BoardStateContext.Provider>
	);
};
