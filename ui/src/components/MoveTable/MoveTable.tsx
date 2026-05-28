import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { getPieceImage, Move, MoveInfo, pieceName, PlayerColors, positionToPGN } from '../../common';
import { useBoardStateContext } from '../../context/BoardStateContext';
import { Message, MessageType, playSound, Sound } from '../../utils';
import styles from './MoveTable.module.css';

interface MoveTableProps {
	mode: 'moves' | 'continuation';
}

export function MoveTable({ mode }: MoveTableProps) {
	const {
		allMoves,
		currentMove,
		settings,
		setSettings,
		displaySettings,
		highlightedMove,
		setHighlightedMove,
		setViewMove,
		sendMessage,
		playContinuationFromCurrent,
	} = useBoardStateContext();

	const movesMode = mode === 'moves';

	const continuation = allMoves[currentMove]?.continuation;
	const latestContinuation = allMoves[allMoves.length - 1]?.continuation;

	const moves: Move[] = useMemo(() => movesMode ? allMoves : (continuation ?? latestContinuation ?? []), [
		movesMode,
		allMoves,
		continuation,
		latestContinuation,
	]);
	const startOffset = movesMode ? 0 : (continuation ? currentMove : allMoves.length - 1);

	const [selectedMove, setSelectedMove] = useState<number>(movesMode ? currentMove : -1);
	const [disableHover, setDisableHover] = useState<boolean>(false);

	const selectMove = useCallback((index: number) => {
		// Stop the engine, set all players to human.
		if (settings.humanPlayers.length !== PlayerColors.length) {
			const newSettings = { ...settings, humanPlayers: PlayerColors.map((_, i) => i) };
			sendMessage(new Message(MessageType.SetSettings, { ...newSettings, evalLimit: newSettings.evalLimit * 1000 }));
			setSettings(newSettings);
		}

		setSelectedMove(index);
		setViewMove(index);
		sendMessage(new Message(MessageType.SetCurrentMove, index));
	}, [settings, setSettings, setViewMove, sendMessage]);

	const playContinuation = useCallback((index: number) => {
		if (index < 1) {
			return;
		}

		const source = continuation ?? latestContinuation;
		if (!source) {
			return;
		}

		playContinuationFromCurrent(source, index);
	}, [continuation, latestContinuation, playContinuationFromCurrent]);

	const handleClick = (index: number) => {
		if (movesMode) {
			setDisableHover(true);
			selectMove(index);
		} else {
			playContinuation(index);
		}

		playSound(Sound.Move);
	};

	const handleMouseEnter = (index: number) => {
		if (movesMode) {
			if (!disableHover) {
				setViewMove(index);
			}
		}

		setHighlightedMove({
			move: moves[index],
			color: PlayerColors[(index + startOffset) % 4],
		});
	};

	const handleMouseLeave = () => {
		if (movesMode) {
			return;
		} else {
			setHighlightedMove(null);
		}
	};

	const handleTableMouseLeave = () => {
		if (movesMode) {
			setDisableHover(false);
			selectMove(selectedMove ?? moves.length - 1);
			setHighlightedMove(null);
		}
	};

	useEffect(() => {
		const handler = (e: KeyboardEvent) => {
			if (movesMode) {
				if (!['ArrowUp', 'ArrowDown'].includes(e.key)) {
					return;
				}

				let newMove: number = -1;

				switch (e.key) {
					case 'ArrowUp':
						newMove = selectedMove - 1;
						break;
					case 'ArrowDown':
						newMove = selectedMove + 1;
						break;
				}

				if (newMove >= 0 && newMove < moves.length) {
					setDisableHover(false);
					selectMove(newMove);

					playSound(Sound.Move);
				}
			} else {
				if (!['ArrowLeft', 'ArrowRight'].includes(e.key)) {
					return;
				}

				if (moves.length === 0) {
					return;
				}

				const currentMove = highlightedMove ? moves.indexOf(highlightedMove.move) : -1;
				let nextMove: number = -1;

				switch (e.key) {
					case 'ArrowRight':
						nextMove = currentMove + 1;
						if (nextMove >= moves.length) {
							return;
						}
						break;
					case 'ArrowLeft':
						nextMove = currentMove === -1 ? moves.length - 1 : currentMove - 1;
						if (nextMove < 0) {
							return;
						}
						break;
				}

				if (nextMove >= 0 && nextMove < moves.length) {
					setHighlightedMove({
						move: moves[nextMove],
						color: PlayerColors[(nextMove + startOffset) % 4],
					});
				}
			}
		};

		window.addEventListener('keydown', handler);
		return () => window.removeEventListener('keydown', handler);
	}, [movesMode, selectedMove, moves, selectMove, highlightedMove, setHighlightedMove, startOffset]);

	if (moves.length === 0) {
		return null;
	}

	const formatMoveDisplay = (move: Move): React.ReactNode => {
		if (move instanceof MoveInfo && move.piece) {
			switch (displaySettings.moveNotation) {
				case 'PGN':
					return <span className={styles.moveText}>{move.toPGN()}</span>;
				case 'SAN':
					return <span className={styles.moveText}>{move.toSAN()}</span>;
				case 'FAN':
				case 'FAN+':
				default:
					const piece = move.piece;
					const captured = move.capturedPiece;

					return captured
						? (
							<>
								<img className={styles.pieceIcon} alt={`${pieceName[piece.type]}`} src={getPieceImage(piece)} />
								<span className={styles.captureX}>x</span>
								{displaySettings.moveNotation === 'FAN+'
									? (
										<img
											className={styles.capturedPieceIcon}
											alt={`${pieceName[captured.type]}`}
											src={getPieceImage(captured)}
										/>
									)
									: null}
								<span className={styles.moveSquare}>{positionToPGN(move.to)}</span>
							</>
						)
						: (
							<>
								<img className={styles.pieceIcon} alt={`${pieceName[piece.type]}`} src={getPieceImage(piece)} />
								<span className={styles.moveSquare}>{positionToPGN(move.to)}</span>
							</>
						);
			}
		}
	};

	// Total cells (including the leading inactive padding cells).
	const totalCells = startOffset + moves.length;
	// Round down the first row to a multiple of 4 so column colors align to players.
	const firstAbsoluteIndex = Math.floor(startOffset / 4) * 4;

	const rows = [];
	for (let i = firstAbsoluteIndex; i < totalCells; i += 4) {
		const cells = Array.from({ length: 4 }).map((_, j) => {
			const cellIndex = i + j;
			const moveIndex = cellIndex - startOffset;
			if (moveIndex < 0 || moveIndex >= moves.length) {
				return <td key={`${i}-${j}-empty`} className={styles.inactiveCell}></td>;
			}

			const isHovered = highlightedMove?.move === moves[moveIndex];
			return (
				<td
					className={[styles.moveCell, moveIndex === selectedMove || isHovered ? styles.currentMove : ''].filter(
						Boolean,
					).join(' ')}
					key={`${i}-${moves[moveIndex].toPGN()}`}
					onClick={() => handleClick(moveIndex)}
					onMouseEnter={() => handleMouseEnter(moveIndex)}
					onMouseLeave={handleMouseLeave}
				>
					{formatMoveDisplay(moves[moveIndex])}
				</td>
			);
		});
		rows.push(
			<tr key={`${i / 4 + 1}-row`}>
				<td className={styles.moveNumber} key={`${i}-number`}>{i / 4 + 1}.</td>
				{cells}
			</tr>,
		);
	}

	return (
		<table className={styles.moveTable} onMouseLeave={handleTableMouseLeave}>
			<thead>{rows}</thead>
		</table>
	);
}
