#! /usr/bin/python3

# rullo.py

# A tool to solve a Rullo board.
import importlib, sys

def vsum(board, col):
    sum = 0
    for row in range(len(board)):
        sum += board[row][col]
    return sum

def explore(board, solution, r, n, horz, vert):
    global solutions
    if r >= len(board): # we may have a solution
        solved = True
        for i in range(len(solution)):
            if vsum(solution, i) != vert[i]:
                solved = False
                break
        if solved:
            solutions = solutions + 1
            print("{0:3d}: {1}".format(solutions, solution))
    else:
        for i in range(1 << n): # exercise all elements
            k = i
            row = []
            for j in range(n):
                row.append(0 if k % 2 == 0 else board[r][j])
                k >>= 1
            # did we solve this row?
            if (sum(row)) == horz[r]:
                solution[r] = row
                # try the next row
                explore(board, solution, r + 1, n, horz, vert)

def solve(board):
    global solutions
    # extract sums from board
    vert = board.pop()  # vertical sums
    horz = []   # horizontal sums
    for r in range(len(board)):
        horz.append(board[r].pop())
    print("board = {}\n horz = {}\n vert = {}".format(board, horz, vert))
    solutions = 0
    explore(board, board[:], 0, len(board[0]), horz, vert)

if __name__ == "__main__":
    # solve(board)
    if len(sys.argv) < 2:
        print("Usage: " + sys.argv[0] + " <board file> [...]")
    else:
        for b in range(1, len(sys.argv)):
            m = importlib.import_module(sys.argv[b])
            solve(m.board)