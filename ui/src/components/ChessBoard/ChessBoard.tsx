import React from 'react';

import { BOARD_SIZE, Color, Move, movesEqual, positionsEqual, positionToPGN } from '../../common';
import { useBoardStateContext } from '../../context/BoardStateContext';
import { ArrowContainer } from '../Arrow';
import { PlayerIndicator } from '../PlayerIndicator';
import { ScoreDisplay } from '../ScoreDisplay';
import { Square } from '../Square';
import styles from './ChessBoard.module.css';

export function ChessBoard() {
	const {
		board,
		activePlayer,
		moves,
		availableMoves,
		selectedSquare,
		movePiece,
		setSelectedSquare,
		displaySettings,
		hoveredMove,
		handleSquareRightMouseDown,
		handleSquareMouseEnter,
		handleSquareRightMouseUp,
		handleSquareLeftClick,
	} = useBoardStateContext();

	const selectedScore = moves[moves.length - 1]?.score ?? null;

	const higlightedSquares: { row: number, col: number, color: Color }[] = [];
	for (let i = moves.length - 1; i >= 0 && i > moves.length - 5; i--) {
		const move = moves[i];
		if (move?.piece) {
			higlightedSquares.push({ ...move.from, color: move.piece.color });
			higlightedSquares.push({ ...move.to, color: move.piece.color });
		}
	}

	if (selectedSquare) {
		higlightedSquares.push({ ...selectedSquare, color: activePlayer });
		higlightedSquares.push(
			...availableMoves.filter(m => positionsEqual(m.from, selectedSquare)).map(m => ({
				...m.to,
				color: activePlayer,
			})),
		);
	}

	if (hoveredMove) {
		higlightedSquares.push({ ...hoveredMove.move.from, color: hoveredMove.color });
		higlightedSquares.push({ ...hoveredMove.move.to, color: hoveredMove.color });
	}

	const handleSquareClick = (row: number, col: number) => {
		const newPosition = { row, col };
		if (board[row][col] === undefined) return;

		// Clear arrows on left click
		handleSquareLeftClick();

		if (selectedSquare) {
			if (movePiece(new Move(selectedSquare, newPosition), true)) {
				setSelectedSquare(null);
			} else if (selectedSquare.col === col && selectedSquare.row === row) {
				setSelectedSquare(null);
			} else if (board[row][col]?.color === activePlayer) {
				setSelectedSquare(newPosition);
			}
		} else if (board[row][col]?.color === activePlayer) {
			setSelectedSquare(newPosition);
		}
	};

	const handleSquareRightClick = (row: number, col: number, event: React.MouseEvent) => {
		event.preventDefault(); // Prevent context menu
	};

	const handleSquareMouseUp = (row: number, col: number, event: React.MouseEvent) => {
		const position = { row, col };
		// Allow arrows on all playable squares, including empty ones
		if (board[row][col] === undefined) return; // Only skip unplayable (corner) squares

		if (event.button === 2) { // Right mouse button
			handleSquareRightMouseUp(position);
		}
	};

	const handleSquareMouseDown = (row: number, col: number, event: React.MouseEvent) => {
		if (event.button === 2) { // Right mouse button
			event.preventDefault();
			const position = { row, col };
			// Allow arrows on all playable squares, including empty ones
			if (board[row][col] === undefined) return; // Only skip unplayable (corner) squares

			handleSquareRightMouseDown(position);
		}
	};

	const handleSquareMouseEnterEvent = (row: number, col: number) => {
		const position = { row, col };
		// Allow arrows on all playable squares, including empty ones
		if (board[row][col] === undefined) return; // Only skip unplayable (corner) squares

		handleSquareMouseEnter(position);
	};

	const getLabel = (row: number, col: number): string => {
		const label = positionToPGN({ row, col });

		switch (displaySettings.showLabels) {
			case 'all':
				return label;
			case 'border':
				for (const [i, j] of [[0, -1], [0, 1], [-1, 0], [1, 0]]) {
					if (board[row + i]?.[col + j] === undefined) {
						return label;
					}
				}
				return '';
			case 'pieces':
				return !!board[row][col] ? label : '';
			case 'moves':
				return !!board[row][col] || higlightedSquares.some(m => positionsEqual(m, { row, col })) ? label : '';
			case 'moves+':
				return !!board[row][col] || availableMoves.some(m => positionsEqual(m.to, { row, col })) ? label : '';
			default:
				return '';
		}
	};

	return (
		<div className={styles.boardContainer}>
			<ScoreDisplay
				score={selectedScore}
				hidden={!displaySettings.showEvalBar}
				showScore={displaySettings.showEvalBarScore}
			/>
			<div className={styles.boardInnerContainer}>
				<div className={styles.board}>
					{Array(BOARD_SIZE).fill(null).map((_, row) => (
						<div
							className={styles.row}
							key={row}
							style={{
								height: `${100 / BOARD_SIZE}%`,
							}}
						>
							{Array(BOARD_SIZE).fill(null).map((_, col) => (
								<Square
									key={`${row}-${col}`}
									isPlayable={board[row][col] !== undefined}
									isLight={(row + col) % 2 === 0}
									piece={board[row][col]}
									higlighted={higlightedSquares.find(square => square.row === row && square.col === col)?.color ?? null}
									possibleMove={!!selectedSquare
										&& availableMoves.some(m => movesEqual(m, new Move(selectedSquare, { row, col })))}
									label={getLabel(row, col)}
									onClick={() => handleSquareClick(row, col)}
									onContextMenu={e => handleSquareRightClick(row, col, e)}
									onMouseDown={e => handleSquareMouseDown(row, col, e)}
									onMouseUp={e => handleSquareMouseUp(row, col, e)}
									onMouseEnter={() => handleSquareMouseEnterEvent(row, col)}
								/>
							))}
						</div>
					))}
					<div className={styles.centerMarker} />

					{/* Arrows */}
					<ArrowContainer />
				</div>
				<PlayerIndicator color={activePlayer} />
			</div>
		</div>
	);
}
