import React from 'react';
import { Color, colorCode } from '../../common';
import { useBoardStateContext } from '../../context/BoardStateContext';
import { MoveNotation, ShowLabels } from '../../utils';
import { moveNotationOptions, showLabelsOptions } from '../../utils/GameStateManager';
import { Checkbox } from '../Checkbox';
import styles from './DisplaySettings.module.css';

function capitalize(s: string): string {
	return s.charAt(0).toUpperCase() + s.slice(1);
}

export function DisplaySettings() {
	const { displaySettings, setDisplaySettings } = useBoardStateContext();

	return (
		<div className={styles.displaySettings}>
			<div className={styles.displaySettingsTable}>
				<table>
					<tbody>
						<tr>
							<td>Move Notation:</td>
							<td>
								<select
									value={displaySettings.moveNotation}
									onChange={e =>
										setDisplaySettings({ ...displaySettings, moveNotation: e.target.value as MoveNotation })}
								>
									{moveNotationOptions.map(option => <option key={option} value={option}>{capitalize(option)}</option>)}
								</select>
							</td>
						</tr>
						<tr>
							<td>Square Labels:</td>
							<td>
								<select
									value={displaySettings.showLabels}
									onChange={e => setDisplaySettings({ ...displaySettings, showLabels: e.target.value as ShowLabels })}
								>
									{showLabelsOptions.map(option => <option key={option} value={option}>{capitalize(option)}</option>)}
								</select>
							</td>
						</tr>
						<tr>
							<td>Show Continuation:</td>
							<td>
								<Checkbox
									checked={displaySettings.showContinuation}
									onChange={checked => setDisplaySettings({ ...displaySettings, showContinuation: checked })}
									background={colorCode(Color.DarkGray)}
									borderColor={colorCode(Color.DarkGray)}
								/>
							</td>
						</tr>
						<tr>
							<td>Show Eval Bar:</td>
							<td>
								<Checkbox
									checked={displaySettings.showEvalBar}
									onChange={checked => setDisplaySettings({ ...displaySettings, showEvalBar: checked })}
									background={colorCode(Color.DarkGray)}
									borderColor={colorCode(Color.DarkGray)}
								/>
							</td>
						</tr>
						<tr>
							<td>Eval Bar Score:</td>
							<td>
								<Checkbox
									checked={displaySettings.showEvalBarScore}
									onChange={checked => setDisplaySettings({ ...displaySettings, showEvalBarScore: checked })}
									background={colorCode(Color.DarkGray)}
									borderColor={colorCode(Color.DarkGray)}
								/>
							</td>
						</tr>
					</tbody>
				</table>
			</div>
		</div>
	);
}
