import { flatbuffers } from "https://cdnjs.cloudflare.com/ajax/libs/flatbuffers/2.0.0/flatbuffers.mjs";
import { Message } from "./message_generated.js";

let ws;

window.onload = () => {
    ws = new WebSocket("ws://localhost:8080/ws");

    ws.onopen = () => {
        console.log("Connected to the WebSocket server");
    };

    ws.onmessage = (event) => {
        const data = new Uint8Array(event.data);
        const buf = new flatbuffers.ByteBuffer(data);
        const message = Message.getRootAsMessage(buf);

        const role = message.user();
        const content = message.content();
        const timestamp = message.timestamp();

        displayMessage(role, content, new Date(timestamp * 1000));
    };

    ws.onclose = () => {
        console.log("Disconnected from the WebSocket server");
    };
};

function sendMessage() {
    const input = document.getElementById("messageInput");
    const content = input.value;
    if (content.trim() === "") return;

    const builder = new flatbuffers.Builder(1024);

    const id = builder.createString("1");
    const user = builder.createString("user");
    const contentOffset = builder.createString(content);
    const timestamp = Math.floor(Date.now() / 1000);

    Message.startMessage(builder);
    Message.addId(builder, id);
    Message.addUser(builder, user);
    Message.addContent(builder, contentOffset);
    Message.addTimestamp(builder, timestamp);
    const message = Message.endMessage(builder);

    builder.finish(message);
    const buf = builder.asUint8Array();

    ws.send(buf);

    displayMessage("user", content, new Date(timestamp * 1000));
    input.value = "";
}

function displayMessage(role, content, timestamp) {
    const messages = document.getElementById("messages");
    const messageElement = document.createElement("div");
    messageElement.className = `message ${role}`;
    messageElement.textContent = `[${timestamp.toLocaleTimeString()}] ${role}: ${content}`;
    messages.appendChild(messageElement);
    messages.scrollTop = messages.scrollHeight;
}
