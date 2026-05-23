import React from 'react';
import { BOARD_SIZE, Color, colorCode, CORNER_SIZE } from '../../common';
import styles from './ScoreDisplay.module.css';

export function ScoreDisplay(
	{ score, hidden, showScore }: { score: number | null, hidden?: boolean, showScore?: boolean },
) {
	const offsetLength = `calc(${CORNER_SIZE / BOARD_SIZE} * 100%)`;
	const height = calculateHeight(score);
	const labelOnBlue = height >= 95;
	const labelText = score === null
		? ''
		: Math.abs(score) > 900
		? `M${1000 - Math.abs(score)}`
		: Math.abs(score).toFixed(1);

	function calculateHeight(score: number | null): number {
		const baseHeight = 50;

		if (score === null) {
			return baseHeight;
		}
		if (score < 0) {
			return 100 - calculateHeight(-score);
		}

		const squareHeight = 100 / 8;

		switch (true) {
			case score < 4:
				return baseHeight + score / 4 * squareHeight;
			case score < 16:
				return baseHeight + (1 + score / 16) * squareHeight;
			case score < 64:
				return baseHeight + (2 + score / 64) * squareHeight;
			case score < 990:
				return baseHeight + 3 * squareHeight;
			case score <= 1000:
				return baseHeight + (3 + (score - 990) / 10) * squareHeight;
			default:
				return 100;
		}
	}

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
