// Fake ims.ssd.sk used by the e2e suite; scenarios are set per test via /__control.
import http from "node:http";

const port = Number(process.env.FAKE_SSD_PORT || 4010);
const SESSION = "e2e-session";

let state = { days: {}, fallback: "empty", requests: [], logins: 0 };

function readBody(req) {
  return new Promise((resolve, reject) => {
    let data = "";
    req.on("data", (chunk) => (data += chunk));
    req.on("end", () => resolve(data));
    req.on("error", reject);
  });
}

function send(res, status, body, headers = {}) {
  res.writeHead(status, { "Content-Type": "application/json", ...headers });
  res.end(typeof body === "string" ? body : JSON.stringify(body));
}

function rowsFor(spec) {
  return spec.map(({ incoming, outgoing }) => ({
    values: ["", 1, incoming, "Real", outgoing, "Real", 0, null, 0, null],
  }));
}

const server = http.createServer(async (req, res) => {
  const body = await readBody(req);

  if (req.url === "/__control/reset" && req.method === "POST") {
    state = { days: {}, fallback: "empty", requests: [], logins: 0 };
    return send(res, 200, {});
  }
  if (req.url === "/__control/scenario" && req.method === "POST") {
    const { days = {}, fallback = "empty" } = JSON.parse(body || "{}");
    state.days = days;
    state.fallback = fallback;
    return send(res, 200, {});
  }
  if (req.url === "/__control/requests" && req.method === "GET") {
    return send(res, 200, { requests: state.requests, logins: state.logins });
  }

  if (req.url === "/api/account/login" && req.method === "POST") {
    const { username, password } = JSON.parse(body || "{}");
    if (username !== "e2e-user" || password !== "e2e-pass") return send(res, 401, "bad credentials");
    state.logins += 1;
    return send(res, 200, {}, { "Set-Cookie": `session=${SESSION}; Path=/; HttpOnly` });
  }

  if (req.url === "/api/consumption-production/profile-data" && req.method === "POST") {
    if (!(req.headers.cookie || "").includes(`session=${SESSION}`)) return send(res, 401, "not logged in");
    const parsed = JSON.parse(body);
    const pod = parsed.filters.find((f) => f.member === "pointOfDeliveryId")?.value;
    const window = parsed.filters.find((f) => f.member === "meteringDatetime" && f.operator === "Greater");
    const date = window.value.slice(0, 10);
    state.requests.push({ date, pod: String(pod), from: window.value, to: window.rangeValue });

    const spec = state.days[date] ?? state.fallback;
    if (spec === "error") return send(res, 500, "upstream exploded");
    if (spec === "empty") return send(res, 200, { rows: [] });
    return send(res, 200, { rows: rowsFor(spec) });
  }

  send(res, 404, "not found");
});

server.listen(port, "127.0.0.1", () => console.log(`fake SSD listening on ${port}`));
