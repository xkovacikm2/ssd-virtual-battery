import { defineConfig } from "@playwright/test";
import { appPort, databaseUrl, ssdPort, ssdUrl } from "./support/env";

const appEnv = {
  DATABASE_URL: databaseUrl,
  PORT: String(appPort),
  IMS_SSD_BASE_URL: ssdUrl,
  IMS_SSD_USERNAME: "e2e-user",
  IMS_SSD_PASSWORD: "e2e-pass",
  IMS_SSD_POINT_OF_DELIVERY_ID: "4242",
  TZ: "UTC",
};

export default defineConfig({
  testDir: "./tests",
  fullyParallel: false,
  workers: 1,
  retries: 0,
  reporter: [["list"]],
  use: {
    baseURL: `http://127.0.0.1:${appPort}`,
    timezoneId: "UTC",
    locale: "sk-SK",
    trace: "retain-on-failure",
    // Canvas rasterization hangs without a usable GPU stack in slim containers.
    launchOptions: { args: ["--disable-gpu", "--disable-software-rasterizer"] },
  },
  webServer: [
    {
      command: "node fake-ssd.mjs",
      url: `http://127.0.0.1:${ssdPort}/__control/requests`,
      env: { FAKE_SSD_PORT: String(ssdPort) },
      reuseExistingServer: false,
    },
    {
      command: "bash scripts/start-go.sh",
      url: `http://127.0.0.1:${appPort}/manifest.webmanifest`,
      env: appEnv,
      timeout: 180_000,
      reuseExistingServer: false,
      stdout: "pipe",
    },
  ],
});
