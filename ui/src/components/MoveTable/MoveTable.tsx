import React, { useState } from 'react';
import { getPieceImage, Move, MoveInfo, pieceName, PlayerColors, positionToPGN } from '../../common';
import { useBoardStateContext } from '../../context/BoardStateContext';
import { MoveNotation, OnMoveHover } from '../../utils';
import styles from './MoveTable.module.css';

interface MoveTableProps {
	moves: Move[];
	currentMove: number;
	moveNotation: MoveNotation;
	handleSetCurrentMove: (moveIndex: number) => void;
	handleSetViewMove?: (moveIndex: number) => void;
	startOffset?: number;
	overrideHoverMode?: OnMoveHover;
}

export function MoveTable(
	{ moves, currentMove, moveNotation, handleSetCurrentMove, handleSetViewMove, startOffset = 0, overrideHoverMode }:
		MoveTableProps,
) {
	const { setHoveredMove, hoveredMove } = useBoardStateContext();
	const [selectedMove, setSelectedMove] = useState<number | null>(null);

	const hoverMode = overrideHoverMode ?? 'set board';
	const onHoverSetBoard = handleSetViewMove ?? handleSetCurrentMove;

	const handleMouseEnter = (moveIndex: number) => {
		let mode = selectedMove ? 'selectedMove' : hoverMode;

		switch (mode) {
			case 'arrow':
			case 'highlight':
			case 'highlight+':
			case 'selectedMove':
				setHoveredMove({ move: moves[moveIndex], color: PlayerColors[(moveIndex + startOffset) % 4] });
				break;
			case 'set board':
				onHoverSetBoard(moveIndex);
				break;
			case 'none':
				setHoveredMove(null);
				break;
		}
	};

	const handleMouseLeave = (moveIndex: number) => {
		switch (hoverMode) {
			case 'arrow':
			case 'highlight':
			case 'highlight+':
				setHoveredMove(null);
				break;
			case 'set board':
				// Handled by handleTableMouseLeave.
				break;
			case 'none':
				break;
		}
	};

	const handleTableMouseLeave = () => {
		if (hoverMode !== 'set board') {
			return;
		}

		if (selectedMove !== null) {
			setHoveredMove(null);
			return;
		}

		const target = moves.length - 1;
		if (target !== currentMove) {
			onHoverSetBoard(target);
		}
	};

	const handleClick = (moveIndex: number) => {
		setSelectedMove(moveIndex);
		handleSetCurrentMove(moveIndex);
		onHoverSetBoard(moveIndex);
	};

	const formatMoveDisplay = (move: Move): React.ReactNode => {
		if (move instanceof MoveInfo && move.piece) {
			switch (moveNotation) {
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
								{moveNotation === 'FAN+'
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

			const isHovered = hoveredMove?.move === moves[moveIndex];
			return (
				<td
					className={[styles.moveCell, moveIndex === currentMove || isHovered ? styles.currentMove : ''].filter(Boolean)
						.join(' ')}
					key={`${i}-${moves[moveIndex].toPGN()}`}
					onClick={() => handleClick(moveIndex)}
					onMouseEnter={() => handleMouseEnter(moveIndex)}
					onMouseLeave={() => handleMouseLeave(moveIndex)}
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
