import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { movesToPGN, PlayerColors } from '../../common';
import { useBoardStateContext } from '../../context/BoardStateContext';
import { Message, MessageType } from '../../utils';
import { CollapsibleBlock } from '../CollapsibleBlock';
import { DisplaySettings } from '../DisplaySettings';
import { MoveTable } from '../MoveTable';
import { Settings } from '../Settings';
import styles from './Menu.module.css';

export function Menu() {
	const {
		allMoves,
		currentMove,
		setCurrentMove,
		setViewMove,
		sendMessage,
		displaySettings,
		hoveredMove,
		setHoveredMove,
		playContinuationFromCurrent,
		settings,
		setSettings,
		connected,
	} = useBoardStateContext();
	const [pgnBlockCollapsed, setPGNBlockCollapsed] = useState(false);
	const [userPGN, setUserPGN] = useState<string | null>(null);

	const pgn = userPGN != null ? userPGN : movesToPGN(allMoves);

	const currentMoveContinuation = allMoves[currentMove]?.continuation;
	const latestContinuation = allMoves[allMoves.length - 1]?.continuation;
	const displayedContinuation = useMemo(() => currentMoveContinuation ?? latestContinuation ?? [], [
		currentMoveContinuation,
		latestContinuation,
	]);
	const continuationAnchor = useMemo(() => currentMoveContinuation ? currentMove : allMoves.length - 1, [
		currentMoveContinuation,
		currentMove,
		allMoves.length,
	]);

	const handleContinuationClick = (continuationIndex: number) => {
		if (continuationIndex < 1) {
			return;
		}
		const source = currentMoveContinuation ?? latestContinuation;
		if (!source) {
			return;
		}
		playContinuationFromCurrent(source, continuationIndex);
	};

	const handleNewGame = (event: React.MouseEvent<HTMLButtonElement>) => {
		sendMessage(new Message(MessageType.NewGame, null));
		event.stopPropagation();
	};

	const handleCopy = (event: React.MouseEvent<HTMLButtonElement>) => {
		navigator.clipboard.writeText(movesToPGN(allMoves));
		setPGNBlockCollapsed(false);
		event.stopPropagation();
	};

	const handleLoad = (event: React.MouseEvent<HTMLButtonElement>) => {
		sendMessage(new Message(MessageType.LoadGame, pgn));
		setUserPGN(null);
		setPGNBlockCollapsed(true);
		event.stopPropagation();
	};

	const handleSetCurrentMove = useCallback((moveIndex: number) => {
		setCurrentMove(moveIndex);
		sendMessage(new Message(MessageType.SetCurrentMove, moveIndex));
	}, [setCurrentMove, sendMessage]);

	const handleSetCurrentMoveFromClick = (moveIndex: number) => {
		if (settings.humanPlayers.length !== PlayerColors.length) {
			const newSettings = {
				...settings,
				humanPlayers: PlayerColors.map((_, i) => i),
			};

			sendMessage(new Message(MessageType.SetSettings, { ...newSettings, evalLimit: newSettings.evalLimit * 1000 }));
			setSettings(newSettings);
		}
		handleSetCurrentMove(moveIndex);
	};

	useEffect(() => {
		const handler = (e: KeyboardEvent) => {
			if (e.key === 'ArrowUp' && currentMove > 0) {
				handleSetCurrentMove(currentMove - 1);
			} else if (e.key === 'ArrowDown' && currentMove < allMoves.length - 1) {
				handleSetCurrentMove(currentMove + 1);
			} else if (e.key === 'ArrowLeft' || e.key === 'ArrowRight') {
				if (displayedContinuation.length === 0) {
					return;
				}
				const cursorIdx = hoveredMove ? displayedContinuation.indexOf(hoveredMove.move) : -1;
				let nextIdx: number;
				if (e.key === 'ArrowRight') {
					nextIdx = cursorIdx + 1;
					if (nextIdx >= displayedContinuation.length) {
						return;
					}
				} else {
					nextIdx = cursorIdx === -1 ? displayedContinuation.length - 1 : cursorIdx - 1;
					if (nextIdx < 0) {
						return;
					}
				}
				setHoveredMove({
					move: displayedContinuation[nextIdx],
					color: PlayerColors[(nextIdx + continuationAnchor) % PlayerColors.length],
				});
			}
		};
		window.addEventListener('keydown', handler);
		return () => window.removeEventListener('keydown', handler);
	}, [
		currentMove,
		allMoves,
		displayedContinuation,
		continuationAnchor,
		hoveredMove,
		setHoveredMove,
		handleSetCurrentMove,
	]);

	return (
		<div className={styles.menuContainer}>
			<div className={styles.menu}>
				<CollapsibleBlock
					collapsed={pgnBlockCollapsed}
					header={
						<div className={styles.buttonGroup}>
							<button id='button-new-game' onClick={handleNewGame}>New Game</button>
							<button id='button-copy' onClick={handleCopy}>Copy</button>
							<button id='button-load' onClick={handleLoad}>Load</button>
							<div
								className={`${styles.connectionIndicator} ${
									connected ? styles.connectionIndicatorConnected : styles.connectionIndicatorDisconnected
								}`}
								title={connected ? 'Connected to the engine' : 'Not connected to the engine'}
							/>
						</div>
					}
				>
					<textarea
						id='game-save-text'
						value={pgn}
						onChange={e => setUserPGN(e.target.value)}
						onBlur={() => setUserPGN(userPGN || movesToPGN(allMoves))} // Reset on empty userPGN.
						className={styles.gameTextarea}
					/>
				</CollapsibleBlock>
				<CollapsibleBlock header='Settings' collapsed={false}>
					<Settings />
				</CollapsibleBlock>
				<CollapsibleBlock header='Display Settings' collapsed={false}>
					<DisplaySettings />
				</CollapsibleBlock>
				<CollapsibleBlock header='Moves' collapsed={false}>
					{allMoves.length > 0 && (
						<MoveTable
							moves={allMoves}
							currentMove={currentMove}
							moveNotation={displaySettings.moveNotation}
							handleSetCurrentMove={handleSetCurrentMoveFromClick}
							handleSetViewMove={setViewMove}
						/>
					)}
				</CollapsibleBlock>
				{displaySettings.showContinuation && (
					<CollapsibleBlock header='Continuation' collapsed={false}>
						{displayedContinuation.length > 0 && (
							<MoveTable
								moves={displayedContinuation}
								currentMove={-1}
								moveNotation={displaySettings.moveNotation}
								handleSetCurrentMove={handleContinuationClick}
								startOffset={continuationAnchor}
								overrideHoverMode='highlight+'
							/>
						)}
					</CollapsibleBlock>
				)}
			</div>
		</div>
	);
}
