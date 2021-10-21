const ws = new WebSocket("wss://" + "tiktaktoe.beprosto1.repl.co" + "/game");

let playerSymbol = "-";

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
  }
};
ws.onclose = () => alert("Game connection lost.");

const acceptBody = document.getElementById("acceptBody");

function fillAcceptBodyWithText() {
  acceptBody.innerHTML = ` END OF THE ROUND! <br> <button id="reset">New Round!</button> `;
}
function requestWaitAcceptBodyText() {
  acceptBody.innerHTML = ` Please wait for the second player... `;
}
function emptyAcceptBodyText() {
  acceptBody.innerHTML = ``;
}

const gridBody = document.getElementById("gridBody");

function putGridElement(x, y) {
  if(window.innerWidth >= 480) {
    gridBody.innerHTML += `<button class="gridElementsNormal" id="gridElement[${x},${y}]" onclick="gridElementHandle(${x}, ${y})"> = </button>`;
  }
  else {
    gridBody.innerHTML += `<button class="gridElementsSmall" id="gridElement[${x},${y}]" onclick="gridElementHandle(${x}, ${y})"> = </button>`;
  }
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

generateGrid(10, 10);