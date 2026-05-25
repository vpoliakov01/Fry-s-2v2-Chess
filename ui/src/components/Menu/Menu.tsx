import React, { useState } from 'react';
import { movesToPGN } from '../../common';
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
		sendMessage,
		displaySettings,
		connected,
	} = useBoardStateContext();
	const [pgnBlockCollapsed, setPGNBlockCollapsed] = useState(false);
	const [userPGN, setUserPGN] = useState<string | null>(null);

	const pgn = userPGN != null ? userPGN : movesToPGN(allMoves);

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
					<MoveTable mode='moves' />
				</CollapsibleBlock>
				{displaySettings.showContinuation && (
					<CollapsibleBlock header='Continuation' collapsed={false}>
						<MoveTable mode='continuation' />
					</CollapsibleBlock>
				)}
			</div>
		</div>
	);
}
