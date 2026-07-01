import ws from "k6/ws";

export const options = {
  vus: 1,
  duration: "10s",
};

export default function () {
  ws.connect("ws://localhost:8080/ws", {}, function (socket) {
    socket.on("open", function () {
      socket.close();
    });
  });
}

