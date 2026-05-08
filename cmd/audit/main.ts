import amqplib, { ChannelModel, type Channel } from "amqplib";

// ── Constants (matching Go messaging package) ────────────────────────
const TASK_EXCHANGE = "task.events";
const USER_EXCHANGE = "user.events";
const ONBOARDING_EXCHANGE = "onboarding.events";
const QUEUE_AUDIT_TASK_EVENTS = "audit.task.events";
const QUEUE_AUDIT_USER_EVENTS = "audit.user.events";
const QUEUE_AUDIT_ONBOARDING_EVENTS = "audit.onboarding.events";

// ── Types ────────────────────────────────────────────────────────────
interface Message {
  type: string;
  payload: unknown;
}

interface LokiStream {
  stream: Record<string, string>;
  values: [string, string][];
}

interface LokiPush {
  streams: LokiStream[];
}

// ── Loki push ────────────────────────────────────────────────────────
async function pushToLoki(
  lokiURL: string,
  eventType: string,
  payload: unknown,
): Promise<void> {
  const line = JSON.stringify({ event_type: eventType, payload });
  const body: LokiPush = {
    streams: [
      {
        stream: { service: "audit", event_type: eventType },
        values: [[`${BigInt(Date.now()) * 1_000_000n}`, line]],
      },
    ],
  };

  const resp = await fetch(`${lokiURL}/loki/api/v1/push`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });

  if (!resp.ok) {
    const text = await resp.text();
    throw new Error(`loki push: status ${resp.status} — ${text}`);
  }
}

// ── RabbitMQ subscribe (mirrors Go Subscribe function) ───────────────
async function subscribe(
  ch: Channel,
  exchange: string,
  queue: string,
  handler: (msg: Message) => Promise<void> | void,
): Promise<void> {
  await ch.assertExchange(exchange, "fanout", { durable: true });
  await ch.assertQueue(queue, { durable: true });
  await ch.bindQueue(queue, exchange, "");

  await ch.consume(
    queue,
    async (delivery) => {
      if (!delivery) return;
      try {
        const msg: Message = JSON.parse(delivery.content.toString());
        await handler(msg);
        ch.ack(delivery);
      } catch (err) {
        console.error(`[audit] unmarshal/handle error on ${queue}:`, err);
        ch.nack(delivery, false, false);
      }
    },
    { noAck: false },
  );
}

// ── Main ─────────────────────────────────────────────────────────────
const amqpURL = process.env.AMQP_URL ?? "amqp://guest:guest@localhost:5672/";
const lokiURL = process.env.LOKI_URL ?? "http://localhost:3100";

let conn: ChannelModel;
try {
  conn = await amqplib.connect(amqpURL);
} catch (err) {
  console.error("rabbit:", err);
  process.exit(1);
}

// Use separate channels per consumer to avoid starvation
const taskCh = await conn.createChannel();
const onboardingCh = await conn.createChannel();

// Subscribe to task events
await subscribe(taskCh, TASK_EXCHANGE, QUEUE_AUDIT_TASK_EVENTS, async (msg) => {
  console.info(`[audit] received event type=${msg.type}`);
  try {
    await pushToLoki(lokiURL, msg.type, msg.payload);
  } catch (err) {
    console.error("[audit] loki push:", err);
  }
});

// Subscribe to onboarding events
await subscribe(
  onboardingCh,
  ONBOARDING_EXCHANGE,
  QUEUE_AUDIT_USER_EVENTS,
  async (msg) => {
    console.info(`[audit] received event type=${msg.type}`);
    try {
      await pushToLoki(lokiURL, msg.type, msg.payload);
    } catch (err) {
      console.error("[audit] loki push:", err);
    }
  },
);

console.info(
  `[audit] service listening exchange=${TASK_EXCHANGE} exchange=${ONBOARDING_EXCHANGE}`,
);

// Graceful shutdown
const shutdown = async () => {
  console.info("[audit] service stopping");
  await taskCh.close();
  await onboardingCh.close();
  await conn.close();
  process.exit(0);
};

process.on("SIGINT", shutdown);
process.on("SIGTERM", shutdown);
