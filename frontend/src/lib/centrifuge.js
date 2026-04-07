import { Centrifuge } from 'centrifuge';

let client = null;
let subscription = null;

export function connectCentrifugo(url, onEvent) {
  if (client) {
    client.disconnect();
  }

  let wsUrl;
  if (!url) {
    const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    wsUrl = `${proto}//${window.location.host}/connection/websocket`;
  } else {
    wsUrl = url.replace(/^http/, 'ws') + '/connection/websocket';
  }
  client = new Centrifuge(wsUrl);

  client.on('connected', (ctx) => {
    console.log('[centrifugo] connected', ctx);
  });
  client.on('disconnected', (ctx) => {
    console.log('[centrifugo] disconnected', ctx);
  });
  client.on('error', (ctx) => {
    console.error('[centrifugo] client error', ctx);
  });

  subscription = client.newSubscription('status');
  subscription.on('publication', (ctx) => {
    onEvent(ctx.data);
  });
  subscription.on('error', (ctx) => {
    console.error('[centrifugo] subscription error', ctx);
  });

  subscription.subscribe();
  client.connect();

  return () => {
    subscription.unsubscribe();
    client.disconnect();
  };
}
