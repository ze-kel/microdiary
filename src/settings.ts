import { db } from './db/database';
import { settings } from './db/schema';
import { eq } from 'drizzle-orm';

import { z } from 'zod';

// Zod schema to object with all defaults
function getDefaults<Schema extends z.AnyZodObject>(schema: Schema) {
  return Object.fromEntries(
    Object.entries(schema.shape).map(([key, value]) => {
      if (value instanceof z.ZodDefault)
        return [key, value._def.defaultValue()];
      return [key, undefined];
    })
  );
}

// Settings schema. There is no validation on DB level.
// Always set default values/
const ZSettings = z.object({
  timezone: z.string().min(1).default('Europe/London'),
  groupMessagesMinutes: z.coerce.number().min(0).max(60).default(0),
});

export type ISettings = z.infer<typeof ZSettings>;

const defaultSettings = getDefaults(ZSettings) as ISettings;

export const getSettings = async (chatId: number) => {
  const unparsed = await db
    .selectDistinct()
    .from(settings)
    .where(eq(settings.chatId, chatId));

  if (!unparsed || !unparsed[0]) return defaultSettings;

  const result = ZSettings.safeParse(unparsed[0].settings);

  if (result.success) {
    return result.data;
  } else {
    console.log(result.error);
  }

  return defaultSettings;
};

export const setSetting = async <K extends keyof ISettings>({
  key,
  value,
  chatId,
}: {
  key: string;
  value: unknown;
  chatId: number;
}) => {
  const s = (await getSettings(chatId)) as Record<string, unknown>;

  if (key in s) {
    s[key] = value;

    const safe = ZSettings.safeParse(s);

    if (!safe.success) {
      throw new Error(safe.error.message);
    }

    const a = await db
      .insert(settings)
      .values({ chatId, settings: safe.data })
      .onConflictDoUpdate({
        target: settings.chatId,
        set: { settings: safe.data },
      });

    console.log(a);

    return safe.data;
  } else {
    throw new Error('Unknown property');
  }
};
