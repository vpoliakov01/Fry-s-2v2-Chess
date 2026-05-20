# 2v2ChessAI

## What is 2v2 Chess?

A variant of chess with 4 players (1 at every side of the board), where the moves are done in a clockwise order and teammates are on opposite sides of the board, i.e. Red and Yellow vs Blue and Green.
![image](https://user-images.githubusercontent.com/53489500/168638482-0886ab3a-a565-452b-9a94-80c3531cb19b.png)

## Why can't you play normal chess?

I can and I do! But 2v2 chess is more dynamic and requires good teamwork and communication, which can be very rewarding! Also, unlike for 2v2 chess, there are already thousands of engines for normal chess.

## What stage is the project in?

Since it's been started only a few days ago, it is in the MVP stage. It plays generally sound moves, but more optimization / testing is to be done.

![image](https://user-images.githubusercontent.com/53489500/169457551-9ab1c224-d676-4c19-ab04-6b76f1828257.png)

An example of a position reached by the engine playing against itself. Pretty similar to the kind of positions reached by human players.

## How does it work?

To pick a move, it uses negamax with alpha-beta pruning to arrive to the most favorable forced position at a specified depth. How favorable a position is is evaluated based on the team's pieces' positions, progression of the game, and number of available moves. It uses [multithreading by running the position evaluation on all availabe CPUs](https://github.com/vpoliakov01/2v2ChessAI/blob/dev/ai/ai.go#L78-L93) (GPU acceleration is planned for the future).

## What is the ELO estimate for this engine

On depth 5, it is around 1700-1800 ELO

## What are the main components that are worth checking out?
* [ai/ai.go](https://github.com/vpoliakov01/2v2ChessAI/blob/main/ai/ai.go)
* [game/game.go](https://github.com/vpoliakov01/2v2ChessAI/blob/main/ai/game.go)
* [game/board.go](https://github.com/vpoliakov01/2v2ChessAI/blob/main/ai/board.go)
* [game/piece.go](https://github.com/vpoliakov01/2v2ChessAI/blob/main/ai/piece.go)
* [game/ in general](https://github.com/vpoliakov01/2v2ChessAI/tree/main/game)
* [dev branch](https://github.com/vpoliakov01/2v2ChessAI/tree/dev)
* [~~PRs~~](https://github.com/vpoliakov01/2v2ChessAI/pulls?q=+)

Some more positions reached by the engine playing itself:

![image](https://user-images.githubusercontent.com/53489500/169458751-f20fe24b-2372-4ced-937b-75d575195e10.png)
![image](https://user-images.githubusercontent.com/53489500/169458772-539fa726-ffde-4f65-abb7-9e5271950d29.png)

## To play against the AI:
`go build -o cmd/ai cmd/main.go && ./cmd/ai`

## TODO:
### UI:
* Add toggle for game / analysis
* Add more settings

### Engine:
* Filter moves returning captures, development moves, and king safety moves
* Support castling
* Support forced calculation for checks
* Test with very sophisticated position evaluation
    * Fully tune piece position strength
    * Incorporate threat / liability

### Other:
* Dockerize (1 for ui, 1 for the engine)
* Update readme
