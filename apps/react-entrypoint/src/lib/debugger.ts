import { Logger } from "wi-console-logger";

let logger: Logger | null = null;

export function getDebugger() {
  if (!logger) {
    logger = new Logger({
      level: "log",
      transform: {
        colors: {
          log: { background: "black", font: "white" },
          warn: { background: "black", font: "orange" },
          error: { background: "black", font: "red" },
        },
      },
    });
  }

  return logger;
}
