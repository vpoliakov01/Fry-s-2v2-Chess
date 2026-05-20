import { useCallback, useEffect, useRef, useState } from 'react';
import { GameSyncService, Message } from '../utils';

export function useGameSocket(onMessage: (message: Message) => void) {
	const wsRef = useRef<WebSocket | null>(null);
	const handlerRef = useRef(onMessage);
	handlerRef.current = onMessage;
	const [connected, setConnected] = useState(false);

	const sendMessage = useCallback((message: Message) => {
		if (wsRef.current?.readyState === WebSocket.OPEN) {
			console.debug(`Sending    ${message.type}`.padEnd(30), message);
			wsRef.current.send(message.json());
			setConnected(true);
			return;
		}

		setTimeout(() => {
			if (wsRef.current?.readyState === WebSocket.OPEN) {
				console.debug(`Sending    ${message.type}`.padEnd(30), message);
				wsRef.current.send(message.json());
				setConnected(true);
			} else {
				setConnected(false);
			}
		}, 10);
	}, []);

	useEffect(() => {
		const wsUrl = window.location.port === '3000'
			? 'ws://localhost:8080/ws'
			: `${window.location.protocol === 'https:' ? 'wss' : 'ws'}://${window.location.host}/ws`;
		const ws = new WebSocket(wsUrl);

		ws.onopen = () => {
			wsRef.current = ws;
			console.log('WS connected');
			GameSyncService.syncWithEngine(ws);
			setConnected(true);
		};

		ws.onmessage = event => {
			const message = JSON.parse(event.data) as Message;
			console.debug(`Received   ${message.type}`.padEnd(30), message);
			setConnected(true);
			handlerRef.current(message);
		};

		ws.onclose = () => {
			if (wsRef.current === ws) {
				wsRef.current = null;
			}
			console.log('WS disconnected');
			setConnected(false);
		};

		ws.onerror = () => {
			setConnected(false);
		};

		return () => ws.close();
	}, []);

	return { sendMessage, connected };
}
