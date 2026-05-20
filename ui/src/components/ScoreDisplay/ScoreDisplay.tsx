import React from 'react';
import { BOARD_SIZE, Color, colorCode, CORNER_SIZE } from '../../common';
import styles from './ScoreDisplay.module.css';

export function ScoreDisplay({ score, hidden }: { score: number, hidden?: boolean }) {
	const maxScore = 10;
	const offsetLength = `calc(${CORNER_SIZE / BOARD_SIZE} * 100%)`;
	const height = Math.max(Math.min(50 + score / maxScore / 2 * 100, 100), 0);

	return (
		<div
			className={styles.scoreDisplay}
			style={{
				top: offsetLength,
				height: `${(BOARD_SIZE - 2 * CORNER_SIZE) / BOARD_SIZE * 100}%`,
				visibility: hidden ? 'hidden' : 'visible',
			}}
		>
			<div
				className={styles.scoreBarBlue}
				style={{
					backgroundColor: colorCode(Color.Blue),
					height: `${100 - height}%`,
				}}
			/>
			<div
				className={styles.scoreBarRed}
				style={{
					backgroundColor: colorCode(Color.Red),
					height: `${height}%`,
				}}
			/>
		</div>
	);
}
