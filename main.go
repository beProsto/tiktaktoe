package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

const BOARD_WIDTH = 10 // Width of the board
const BOARD_HEIGHT = 10 // Height of the board
const GAME_WIDTH = 5 // How many symbols does there have to be in a line for a player to win

var Board [BOARD_WIDTH * BOARD_HEIGHT]byte

func getBoardElement(x int, y int) byte {
  if(x >= 0 && y >= 0 && x < BOARD_WIDTH && y < BOARD_HEIGHT) {
    return Board[y*BOARD_WIDTH + x]
  } else {
    return 0
  }
}
func setBoardElement(x int, y int, v byte) bool {
  if(x >= 0 && y >= 0 && x < BOARD_WIDTH && y < BOARD_HEIGHT) {
    Board[y*BOARD_WIDTH + x] = v
    return true
  } else {
    return false
  }
}

func processBoard() byte {
  for y := 0; y < BOARD_WIDTH; y++ {
    for x := 0; x < BOARD_WIDTH; x++ {
      var xAxisLineLastCheck bool = true
      var xAxisLineLastSymbol byte = 0
      
      var yAxisLineLastCheck bool = true
      var yAxisLineLastSymbol byte = 0

      var xYAxisLineLastCheck bool = true
      var xYAxisLineLastSymbol byte = 0

      var invXYAxisLineLastCheck bool = true
      var invXYAxisLineLastSymbol byte = 0

      for r := 0; r < GAME_WIDTH; r++ {
        if(r == 0) {
          xAxisLineLastSymbol = getBoardElement(x, y)
          yAxisLineLastSymbol = xAxisLineLastSymbol
          xYAxisLineLastSymbol = xAxisLineLastSymbol
          invXYAxisLineLastSymbol = xAxisLineLastSymbol
        }

        xAxisLineLastCheck = (getBoardElement(x+r, y) == xAxisLineLastSymbol) && xAxisLineLastCheck
        xAxisLineLastSymbol = getBoardElement(x+r, y)

        yAxisLineLastCheck = (getBoardElement(x, y+r) == yAxisLineLastSymbol) && yAxisLineLastCheck
        yAxisLineLastSymbol = getBoardElement(x, y+r)

        xYAxisLineLastCheck = (getBoardElement(x+r, y+r) == xYAxisLineLastSymbol) && xYAxisLineLastCheck
        xYAxisLineLastSymbol = getBoardElement(x+r, y+r)

        invXYAxisLineLastCheck = (getBoardElement(x-r, y+r) == invXYAxisLineLastSymbol) && invXYAxisLineLastCheck
        invXYAxisLineLastSymbol = getBoardElement(x-r, y+r)

        if(r == GAME_WIDTH - 1) {
          if((getBoardElement(x+r, y) == xAxisLineLastSymbol) && xAxisLineLastCheck && xAxisLineLastSymbol != 0) {
            return xAxisLineLastSymbol
          } else if((getBoardElement(x, y+r) == yAxisLineLastSymbol) && yAxisLineLastCheck && yAxisLineLastSymbol != 0) {
            return yAxisLineLastSymbol
          } else if((getBoardElement(x+r, y+r) == xYAxisLineLastSymbol) && xYAxisLineLastCheck && xYAxisLineLastSymbol != 0) {
            return xYAxisLineLastSymbol
          } else if((getBoardElement(x-r, y+r) == invXYAxisLineLastSymbol) && invXYAxisLineLastCheck && invXYAxisLineLastSymbol != 0) {
            return invXYAxisLineLastSymbol
          }
        }
      }
    }
  }

  return 0
}

type ClientData struct {
	symbol byte
}

var XMissing bool = true
var OMissing bool = true

var GameClients sync.Map
var GameClientData sync.Map

var SymbolsTurn byte = 'X'

var XAcceptedEnd bool = false
var OAcceptedEnd bool = false
var RoundEnded   bool = false

func restart() {
// Send Data to all clients
	GameClients.Range(func(key interface{}, value interface{}) bool {
    err := value.(*websocket.Conn).WriteMessage(1, []byte("#RESET"))
		if err != nil {
			fmt.Println("write:", err)
		}

		return true
	})

  // reset the board
  for i := 0; i < BOARD_WIDTH * BOARD_HEIGHT; i++ {
    Board[i] = 0;
  }
}

func game(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Upgrade: ", err)
		return
	}
	defer c.Close()
	fmt.Println("User connected from: ", c.RemoteAddr())

	var clientData ClientData
	if XMissing {
		clientData.symbol = 'X'
		XMissing = false
	} else if OMissing {
		clientData.symbol = 'O'
		OMissing = false
	} else {
		clientData.symbol = '-'
	}
	c.WriteMessage(1, []byte{'&', clientData.symbol})
	GameClientData.Store(c.RemoteAddr(), clientData)

	GameClients.Store(c.RemoteAddr(), c)

  // Send the current board's state to the player
  for i := 0; i < BOARD_WIDTH * BOARD_HEIGHT; i++ {
    x := i%BOARD_WIDTH;
    y := i/BOARD_HEIGHT;
    if(getBoardElement(x, y) != 0) {
      stringToSend := "^" + strconv.Itoa(x) + ":" + strconv.Itoa(y) + ":" + string([]byte{getBoardElement(x, y)})
      err = c.WriteMessage(1, []byte(stringToSend))
		  if err != nil {
				fmt.Println("write:", err)
			}
    }
  }

	for {
		_, message, err2 := c.ReadMessage() //ReadMessage blocks until message received
		msgString := string(message)

		if err2 != nil {
			fmt.Println("read:", err2)
			if clientData.symbol == 'X' {
				XMissing = true
        if(RoundEnded) {
          XAcceptedEnd = true
        }
			} else if clientData.symbol == 'O' {
				OMissing = true
        if(RoundEnded) {
          OAcceptedEnd = true
        }
			}
      GameClientData.Delete(c.RemoteAddr())
			GameClients.Delete(c.RemoteAddr())

      if(XAcceptedEnd && OAcceptedEnd) {
        XAcceptedEnd = false
        OAcceptedEnd = false
        RoundEnded = false
        restart()
      }

			return
		}

		fmt.Println(msgString)

		if msgString != "" && !RoundEnded {
			if msgString[0] == '^' {
				data, ok := GameClientData.Load(c.RemoteAddr())
				msgData := msgString[1:]
				if msgData != "" && ok {
					if data.(ClientData).symbol == SymbolsTurn { // If it is the player's move, change the player and send the change to all the players.
						placement := strings.Split(msgData, ":")
						if len(placement) >= 2 {
							xPlacement, err1 := strconv.ParseUint(placement[0], 10, 8)
							yPlacement, err2 := strconv.ParseUint(placement[1], 10, 8)
							if err1 != nil || err2 != nil || xPlacement >= BOARD_WIDTH || xPlacement < 0 || yPlacement >= BOARD_HEIGHT || yPlacement < 0 {
								fmt.Println("Error accured upon parsing the client input!")
                if xPlacement > BOARD_WIDTH || xPlacement < 0 || yPlacement > BOARD_HEIGHT || yPlacement < 0 {
									fmt.Println("Error had to do with the placement being out of bounds")
								}

							} else if(getBoardElement(int(xPlacement), int(yPlacement)) == 0) {
                // Toggle the symbol between X and O
                if SymbolsTurn == 'X' {
                  SymbolsTurn = 'O'
                } else {
                  SymbolsTurn = 'X'
                }
                
								fmt.Println("X: ", xPlacement, ", Y: ", yPlacement)

								playerSymbol := string([]byte{data.(ClientData).symbol})
								fmt.Println("Symbol: ", playerSymbol)

								stringToSend := "^" + placement[0] + ":" + placement[1] + ":" + playerSymbol
								fmt.Println("Final string: ", stringToSend)

                setBoardElement(int(xPlacement), int(yPlacement), data.(ClientData).symbol)
                fmt.Println(Board)

                playerThatWon := processBoard()
                if(playerThatWon != 0) {
                  RoundEnded = true;
                }

								// Send Data to all clients
								GameClients.Range(func(key interface{}, value interface{}) bool {
									// if value.(*websocket.Conn) == nil {
									// 	GameClients.Delete(key)
									// 	return true
									// }

									err = value.(*websocket.Conn).WriteMessage(1, []byte(stringToSend))
									if err != nil {
										fmt.Println("write:", err)
									}

									clientsData, ok := GameClientData.Load(key)
									if ok {
										if clientsData.(ClientData).symbol == SymbolsTurn {
											err = value.(*websocket.Conn).WriteMessage(1, []byte("@Your Turn."))
											if err != nil {
												fmt.Println("write:", err)
											}
										}
									}

                  if(playerThatWon != 0) {
                    err = value.(*websocket.Conn).WriteMessage(1, []byte("!PLAYER " + string([]byte{playerThatWon}) + " WON!"))
										if err != nil {
											fmt.Println("write:", err)
										}
                    clientsData, ok := GameClientData.Load(key)
                    if(ok && clientsData.(ClientData).symbol != '-') {
                      err = value.(*websocket.Conn).WriteMessage(1, []byte("#REQUEST_RESET_ACCEPT"))
                      if err != nil {
                        fmt.Println("write:", err)
                      }
                    }
                  }

									return true
								})
							}
						}
					} else { // If it's not the player's turn, tell them about it.
						if data.(ClientData).symbol != '-' {
							c.WriteMessage(1, []byte("!Wait for your turn!"))
							if err != nil {
								fmt.Println("write:", err)
							}
						} else {
							c.WriteMessage(1, []byte("!You're not a player!"))
							if err != nil {
								fmt.Println("write:", err)
							}
						}
					}
				} else {
					fmt.Println("Error accured upon loading client data!")
				}
			}
		} else if RoundEnded && msgString == "#READY" && clientData.symbol != '-' {
      if(clientData.symbol == 'X') {
        XAcceptedEnd = true
      } else if(clientData.symbol == 'O') {
        OAcceptedEnd = true
      }

      if(XAcceptedEnd && OAcceptedEnd) {
        XAcceptedEnd = false
        OAcceptedEnd = false
        RoundEnded = false
        restart()
      }
    }
	}
}

func main() {
	fmt.Println("http server up!")

	http.HandleFunc("/game", game)

	fileServer := http.FileServer(http.Dir("./public/"))
	http.Handle("/", fileServer)

	http.ListenAndServe(":0", nil)
}
