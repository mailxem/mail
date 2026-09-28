export interface LoggerOptions {
  fileName: string;
  emoji: string;
  action: string;
  label: string;
  value: Record<string, any> | string | number;
  message: string;
}

enum LogLevel {
  INFO = "INFO",
  WARN = "WARN",
  ERROR = "ERROR",
  DEBUG = "DEBUG"
}

class Logger {
  private getCallerInfo(): { functionName: string; lineNumber: number } {
    const error = new Error();
    const stackLines = error.stack?.split('\n') || [];
    // Skip first 3 lines (Error, getCallerInfo, log method)
    const callerLine = stackLines[3] || '';
    
    // Extract function name and line number
    const functionMatch = callerLine.match(/at\s+(\w+)\s+\(/);
    const lineMatch = callerLine.match(/:(\d+):\d+\)/);
    
    return {
      functionName: functionMatch?.[1] || 'anonymous',
      lineNumber: lineMatch ? parseInt(lineMatch[1], 10) : 0
    };
  }

  private formatLogMessage(options: LoggerOptions, level: LogLevel): string {
    const { fileName, emoji, action, label, value, message } = options;
    const { functionName, lineNumber } = this.getCallerInfo();

    return `[${level}] ${emoji} ${fileName}:${lineNumber} ${functionName}() - [${action}][${label}] ${message} ${
      typeof value === 'string' ? value : JSON.stringify(value)
    }`;
  }

  info(options: LoggerOptions): void {
    const formattedMessage = this.formatLogMessage(options, LogLevel.INFO);
    console.log(formattedMessage);
  }

  warn(options: LoggerOptions): void {
    const formattedMessage = this.formatLogMessage(options, LogLevel.WARN);
    console.warn(formattedMessage);
  }

  error(options: LoggerOptions): void {
    const formattedMessage = this.formatLogMessage(options, LogLevel.ERROR);
    console.error(formattedMessage);
  }

  debug(options: LoggerOptions): void {
    const formattedMessage = this.formatLogMessage(options, LogLevel.DEBUG);
    console.debug(formattedMessage);
  }
}

export const logger = new Logger();