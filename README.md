# 2v2ChessAI

## What is 2v2 Chess?

A variant of chess with 4 players (1 at every side of the board), where the moves are done in a clockwise order and teammates are on opposite sides of the board, i.e. Red and Yellow vs Blue and Green.
![image](https://user-images.githubusercontent.com/53489500/168638482-0886ab3a-a565-452b-9a94-80c3531cb19b.png)

## Why can't you play normal chess?

I can and I do! But 2v2 chess is more dynamic and requires good teamwork and communication, which can be very rewarding! Also, unlike for 2v2 chess, there are already thousands of engines for normal chess.

## What stage is the project in?

Most of the time, it plays very sensible moves, but it's not perfect.

<img width="1990" height="1440" alt="image" src="https://github.com/user-attachments/assets/fc57fddd-d621-4e1b-9f56-3a7ecfc69a8f" />

An example of a position reached by the engine playing against itself. Pretty similar to the kind of positions reached by human players.

## How does it work?

To pick a move, it uses negamax with alpha-beta pruning to arrive to the most favorable position at a specified depth. To find the best move, it checks all possible moves available to the active player (without pruning) concurrently, updating the shared alpha value. For following moves, it heavily prunes the number of moves by running static evalution on them based on the improvement in the piece's position, value of the captured piece (when applicable), threat to the opposing kings, safety of the ally kings, and pawn structure. Evaluations of checked positions are cached. The cache is persisted on disk for moves 1-12 of the game.

## What is the ELO estimate for this engine

On depth 12, spread 8, it is around 1800-2000 ELO

Some more positions reached by the engine playing itself:

<img width="1475" height="1440" alt="image" src="https://github.com/user-attachments/assets/55d980c1-6b66-42cb-aee1-935c0b52b9fb" />

<img width="1476" height="1438" alt="image" src="https://github.com/user-attachments/assets/3b57af64-7a57-4518-94b6-93d204c6469a" />

<img width="1475" height="1438" alt="image" src="https://github.com/user-attachments/assets/3be7ec37-ab78-49dd-825b-87f1abded664" />

## To launch:
1. `docker compose up --build`
2. Open http://localhost
This starts the engine and the UI. (set `UI_PORT` to use a different port, e.g. `UI_PORT=3000 docker compose up --build`).
The engine's cache is persisted in the `chess-cache` volume between restarts.

## TODO:
### UI:
* Move animation (?)
* Thinking indicator

### Engine:
* Run calculation in the background
* Stream currently best discovered lines
* Forced calculation for checks
