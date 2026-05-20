import React from 'react';
import { BOARD_SIZE, Color, colorCode, CORNER_SIZE } from '../../common';
import styles from './ScoreDisplay.module.css';

export function ScoreDisplay({ score, hidden, showScore }: { score: number, hidden?: boolean, showScore?: boolean }) {
	const maxScore = 10;
	const offsetLength = `calc(${CORNER_SIZE / BOARD_SIZE} * 100%)`;
	const height = Math.max(Math.min(50 + score / maxScore / 2 * 100, 100), 0);
	const labelOnBlue = height >= 95;
	const labelText = Math.abs(score) > 900 ? `M${1000 - Math.abs(score)}` : `${score}`;

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
			{showScore && (
				<div
					className={styles.scoreLabel}
					style={{
						top: `calc(${100 - height}% + ${(labelOnBlue ? 1 : -1) * 2}px)`,
						transform: labelOnBlue ? 'none' : 'translateY(-100%)',
						color: colorCode(Color.Black),
					}}
				>
					{labelText}
				</div>
			)}
		</div>
	);
}
