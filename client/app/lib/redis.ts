import { Redis } from 'ioredis';
import {logger} from '@/app/lib/logger';

const FILE_NAME = 'lib/redis.ts';

class RedisClient {
  private static instance: Redis;
  
  private constructor() {}

  public static getInstance(): Redis {
    if (!RedisClient.instance) {
      logger.info({
        fileName: FILE_NAME,
        label: 'Redis',
        value: RedisClient.instance,
        emoji: '🔌',
        action: 'getInstance',
        message: 'Creating new Redis connection'
      });
      
      RedisClient.instance = new Redis(process.env.REDIS_URL);

      RedisClient.instance.on('connect', () => {
        logger.info({
          fileName: FILE_NAME, 
          label: 'Redis',
          emoji: '🔌',
          action: 'connect',
          value: RedisClient.instance,
          message: '✅ Redis connected successfully'
        });
      });

      RedisClient.instance.on('error', (error) => {
        logger.error({
          fileName: FILE_NAME,
          label: 'Redis',
          emoji: '🔌',
          action: 'error',
          value: error,
          message: `❌ Redis connection error, error:${error.message}`
        });
      });
    }

    return RedisClient.instance;
  }
}

// Export a singleton instance
export const redis = RedisClient.getInstance();

// For testing purposes
export const disconnectRedis = async () => {
  if (RedisClient.getInstance()) {
    logger.info({
      fileName: FILE_NAME,
      label: 'Redis', 
      value: RedisClient.getInstance(),
      emoji: '🔌',
      action: 'disconnect',
      message: '🔌 Disconnecting Redis'
    });
    await RedisClient.getInstance().quit();
  }
};
