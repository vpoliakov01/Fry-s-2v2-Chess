export enum Color {
	Red = 'red',
	Blue = 'blue',
	Yellow = 'yellow',
	Green = 'green',
	Black = 'black',
	LightGray = 'light-gray',
	Gray = 'gray',
	DarkGray = 'dark-gray',
	DarkerGray = 'darker-gray',
	White = 'white',
}

export const colorCode = (color: Color) => `var(--color-${color})`;

export const PlayerColors = [Color.Red, Color.Blue, Color.Yellow, Color.Green];
