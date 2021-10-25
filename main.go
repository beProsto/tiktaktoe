package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"strconv"
	"strings"
	"sync"
  "github.com/beProsto/tiktaktoe/randstr" // this is by far the stupidest thing about this language, i hate this so much, 
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

const BOARD_WIDTH = 10 // Width of the board
const BOARD_HEIGHT = 10 // Height of the board
const GAME_WIDTH = 5 // How many symbols does there have to be in a line for a player to win

func restart(board *[BOARD_WIDTH*BOARD_HEIGHT]byte) {
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
    board[i] = 0;
  }
}
func getBoardElement(board *[BOARD_WIDTH*BOARD_HEIGHT]byte, x int, y int) byte {
  if(x >= 0 && y >= 0 && x < BOARD_WIDTH && y < BOARD_HEIGHT) {
    return board[y*BOARD_WIDTH + x]
  } else {
    return 0
  }
}
func setBoardElement(board *[BOARD_WIDTH*BOARD_HEIGHT]byte, x int, y int, v byte) bool {
  if(x >= 0 && y >= 0 && x < BOARD_WIDTH && y < BOARD_HEIGHT) {
    board[y*BOARD_WIDTH + x] = v
    return true
  } else {
    return false
  }
}

func processBoard(board *[BOARD_WIDTH*BOARD_HEIGHT]byte) byte {
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
          xAxisLineLastSymbol = getBoardElement(board, x, y)
          yAxisLineLastSymbol = xAxisLineLastSymbol
          xYAxisLineLastSymbol = xAxisLineLastSymbol
          invXYAxisLineLastSymbol = xAxisLineLastSymbol
        }

        xAxisLineLastCheck = (getBoardElement(board, x+r, y) == xAxisLineLastSymbol) && xAxisLineLastCheck
        xAxisLineLastSymbol = getBoardElement(board, x+r, y)

        yAxisLineLastCheck = (getBoardElement(board, x, y+r) == yAxisLineLastSymbol) && yAxisLineLastCheck
        yAxisLineLastSymbol = getBoardElement(board, x, y+r)

        xYAxisLineLastCheck = (getBoardElement(board, x+r, y+r) == xYAxisLineLastSymbol) && xYAxisLineLastCheck
        xYAxisLineLastSymbol = getBoardElement(board, x+r, y+r)

        invXYAxisLineLastCheck = (getBoardElement(board, x-r, y+r) == invXYAxisLineLastSymbol) && invXYAxisLineLastCheck
        invXYAxisLineLastSymbol = getBoardElement(board, x-r, y+r)

        if(r == GAME_WIDTH - 1) {
          if((getBoardElement(board, x+r, y) == xAxisLineLastSymbol) && xAxisLineLastCheck && xAxisLineLastSymbol != 0) {
            return xAxisLineLastSymbol
          } else if((getBoardElement(board, x, y+r) == yAxisLineLastSymbol) && yAxisLineLastCheck && yAxisLineLastSymbol != 0) {
            return yAxisLineLastSymbol
          } else if((getBoardElement(board, x+r, y+r) == xYAxisLineLastSymbol) && xYAxisLineLastCheck && xYAxisLineLastSymbol != 0) {
            return xYAxisLineLastSymbol
          } else if((getBoardElement(board, x-r, y+r) == invXYAxisLineLastSymbol) && invXYAxisLineLastCheck && invXYAxisLineLastSymbol != 0) {
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
  roomId string
}

type RoomData struct {
  XMissing bool
  OMissing bool
  SymbolsTurn byte
  XAcceptedEnd bool
  OAcceptedEnd bool
  RoundEnded   bool
  Board [BOARD_WIDTH * BOARD_HEIGHT]byte
}
func newRoomData() RoomData {
  return RoomData{
    XMissing: true,
    OMissing: true,
    SymbolsTurn: 'X',
    XAcceptedEnd: false,
    OAcceptedEnd: false,
  };
}

var Rooms sync.Map
var GameClients sync.Map
var GameClientData sync.Map

func makeNewRoom(c *websocket.Conn) string {
  roomId := randstr.StringWithCharset(6, "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
  _, loaded := Rooms.Load(roomId)
  if !loaded {
    err := c.WriteMessage(1, []byte("+" + roomId))
    if err != nil {
      fmt.Println("write:", err)
    }
    Rooms.Store(roomId, &newRoomData())
    return roomId;
  } else {
    return makeNewRoom(c)
  }
}

func connectPlayerToRoom(c *websocket.Conn, cd *ClientData, roomId string) {
  room, exists := Rooms.Load(roomId)
  if exists {
    // here we need client connection code
    cd.roomId = roomId
    // here we send the client info for them to start the game
    err := c.WriteMessage(1, []byte("#START"))
    if err != nil {
      fmt.Println("write:", err)
    }
    room.

  } else {
    err := c.WriteMessage(1, []byte("#WRONG"))
    if err != nil {
      fmt.Println("write:", err)
    }
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

  var roomId string
	var clientData *ClientData = &ClientData{}
	// if XMissing {
	// 	clientData.symbol = 'X'
	// 	XMissing = false
	// } else if OMissing {
	// 	clientData.symbol = 'O'
	// 	OMissing = false
	// } else {
	// 	clientData.symbol = '-'
	// }
	// c.WriteMessage(1, []byte{'&', clientData.symbol})

  // Send the current board's state to the player
  // for i := 0; i < BOARD_WIDTH * BOARD_HEIGHT; i++ {
  //   x := i%BOARD_WIDTH;
  //   y := i/BOARD_HEIGHT;
  //   if(getBoardElement(x, y) != 0) {
  //     stringToSend := "^" + strconv.Itoa(x) + ":" + strconv.Itoa(y) + ":" + string([]byte{getBoardElement(x, y)})
  //     err = c.WriteMessage(1, []byte(stringToSend))
	// 	  if err != nil {
	// 			fmt.Println("write:", err)
	// 		}
  //   }
  // }

	GameClientData.Store(c.RemoteAddr(), clientData)
	GameClients.Store(c.RemoteAddr(), c)

	for {
		_, message, err2 := c.ReadMessage() //ReadMessage blocks until message received
		msgString := string(message)

		if err2 != nil {
			fmt.Println("read:", err2)

			// if clientData.symbol == 'X' {
			// 	XMissing = true
      //   if(RoundEnded) {
      //     XAcceptedEnd = true
      //   }
			// } else if clientData.symbol == 'O' {
			// 	OMissing = true
      //   if(RoundEnded) {
      //     OAcceptedEnd = true
      //   }
			// }
      // GameClientData.Delete(c.RemoteAddr())
			// GameClients.Delete(c.RemoteAddr())

      // if(XAcceptedEnd && OAcceptedEnd) {
      //   XAcceptedEnd = false 
      //   OAcceptedEnd = false
      //   RoundEnded = false
      //   restart()
      // }

			return
		}

		fmt.Println(msgString)

    if msgString == "+" { // When a player wants to create a room
      roomId = makeNewRoom(c)

      // here we need client connection code
      clientData.roomId = roomId 
    } else if msgString[0] == '%' { // When a player wants to connect to a room
      roomIdWanted := msgString[1:]
      println(roomIdWanted)


    } else if msgString != "" && !RoundEnded {
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
							err = c.WriteMessage(1, []byte("!Wait for your turn!"))
							if err != nil {
								fmt.Println("write:", err)
							}
						} else {
							err = c.WriteMessage(1, []byte("!You're not a player!"))
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
