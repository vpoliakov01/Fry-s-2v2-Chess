import { BOARD_SIZE } from './constants';
import { Piece, pieceFANCharacter, pieceSANCharacter } from './pieces';
import { Position, positionToPGN } from './positions';

export type PGNMove = string;

export class Move {
	public continuation?: Move[];

	constructor(public from: Position, public to: Position) {}

	static fromPGN(pgn: PGNMove): Move {
		const [from, to] = pgn.split('-').map(pos => ({
			col: pos.charCodeAt(0) - 'a'.charCodeAt(0),
			row: BOARD_SIZE - parseInt(pos.slice(1)),
		}));
		return new Move(from, to);
	}

	toPGN(): PGNMove {
		const [from, to] = [this.from, this.to].map(pos => positionToPGN(pos));
		return `${from}-${to}`;
	}
}

export class MoveInfo extends Move {
	constructor(
		public from: Position,
		public to: Position,
		public piece: Piece,
		public capturedPiece: Piece | null,
		public score: number | null = null,
	) {
		super(from, to);
	}

	toSAN(): string {
		let pieceChar = pieceSANCharacter[this.piece.type];
		const captureChar = this.capturedPiece ? 'x ' : '';
		return `${pieceChar}${captureChar}${positionToPGN(this.to)}`;
	}

	toFAN(): string {
		let pieceChar = pieceFANCharacter[this.piece.type];
		const captureChar = this.capturedPiece ? 'x' : '';
		return `${pieceChar}${captureChar}${positionToPGN(this.to)}`;
	}
}

export function movesEqual(a: Move, b: Move): boolean {
	return a.from.row === b.from.row && a.from.col === b.from.col && a.to.row === b.to.row && a.to.col === b.to.col;
}

export function movesToPGN(moves: Move[]): string {
	let pgn = '';

	for (let i = 0; i < moves.length; i += 4) {
		if (i > 0 && i % 4 === 0) {
			pgn += '\n';
		}
		pgn += `${i / 4 + 1}.`;
		for (let j = 0; j < 4 && i + j < moves.length; j++) {
			pgn += ` ${moves[i + j].toPGN()}`;
		}
	}

	return pgn;
}
