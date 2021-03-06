// Way of communication with the server
const ws = new WebSocket("wss://" + "tiktaktoe.beprosto1.repl.co" + "/game");

let playerSymbol = "-";

// DOM Structure

// Box that contains information about the game state (wether needs a reset or not)
const acceptBody = document.getElementById("acceptBody");

function fillAcceptBodyWithText() {
  acceptBody.style.display = "var(--visible)";
  acceptBody.innerHTML = ` END OF THE ROUND! <br> <button id="reset">New Round!</button> `;
}
function requestWaitAcceptBodyText() {
  acceptBody.innerHTML = ` Please wait for the second player... `;
}
function emptyAcceptBodyText() {
  acceptBody.innerHTML = ``;
  acceptBody.style.display = "var(--invisible)";
}

// The game grid itself
const gridBody = document.getElementById("gridBody");

function putGridElement(x, y) {
  gridBody.innerHTML += `<button class="gridElements" id="gridElement[${x},${y}]" onclick="gridElementHandle(${x}, ${y})"> = </button>`;
}
function putGridLineSeparator() {
  gridBody.innerHTML += `<br>`;
}

function setGridElement(x, y, str) {
  document.getElementById(`gridElement[${x},${y}]`).innerHTML = str;
}

function gridElementHandle(x, y) {
  ws.send(`^${x}:${y}`);
}

function generateGrid(sizex, sizey) {
  for(let i = 0; i < sizey; i++) {
    for(let j = 0; j < sizex; j++) {
      putGridElement(j, i);
    }
    putGridLineSeparator();
  }
}

// Room selection screen
const roomSelectionBody = document.getElementById("roomSelectionBody");

const idInput = document.getElementById("idInput");

const connectButton = document.getElementById("connectButton");

const createButton = document.getElementById("createButton");

connectButton.onclick = () => {
  if(idInput.value.length == 6) {
    ws.send("%" + idInput.value);
  }
};

createButton.onclick = () => {
  ws.send("+");
};

// Interpreting the messages from the server
ws.onmessage = (msg) => {
  const text = msg.data;
  if(text[0] == "&") { // Sent when the player connects, assignes a symbol to them
    if(text[1] == "X") {
      console.log("X Connected.");
    }
    else if(text[1] == "O") {
      console.log("O Connected");
    }
    else {
      console.log("Spectator connected")
    }

    playerSymbol = text[1];
    if(playerSymbol != "-") {
      alert(`You are an ${playerSymbol}!`);
    }
    else {
      alert(`You are a Spectator!`)
    }
  }
  else if(text[0] == "^") { // sent every time a move by any player is made
    const toInterpret = text.substring(1, text.length);
    const xys = toInterpret.split(":");
    setGridElement(xys[0], xys[1], xys[2]);
  }
  else if(text[0] == "!") { // sent every time the server warns the player about something
    const message = text.substring(1, text.length);
    if(message.length > 0) {
      alert(message);
    }
  }
  else if(text[0] == "@") { // sent every time the server says something to the player
    const message = text.substring(1, text.length);
    if(message.length > 0) {
      console.log(message);
    }
  }
  else if(text[0] == "#") { // sent every time the server communicates a game state
    const message = text.substring(1, text.length);
    if(message == "RESET") {
      for(let element of document.getElementsByClassName("gridElements")) {
        element.innerHTML = " = ";
      }
      emptyAcceptBodyText();
    }
    else if(message == "REQUEST_RESET_ACCEPT") {
      fillAcceptBodyWithText();
      document.getElementById("reset").onclick = () => {
        requestWaitAcceptBodyText();
        ws.send("#READY");
      };
    }
    else if(message == "START") {
      roomSelectionBody.style.display = "var(--invisible)";
      gridBody.style.display = "var(--visible)";
      generateGrid(10,10);
    }
    else if(message == "WRONG") {
      alert("Room ID Invalid!");
      idInput.value = "";
    }
  }
  else if(text[0] == "+") {
    const id = text.substring(1, text.length);
    alert(id);
  }
};

ws.onclose = () => alert("Game connection lost.");
