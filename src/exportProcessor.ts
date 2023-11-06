import { getSettings } from './settings';
import { IMessage } from './db/schema';

type FormattedMessageDates = {
  day: string;
  month: string;
  time: string;
  date: Date;
};

const getDates = (
  message: IMessage,
  timeZone: string
): FormattedMessageDates => {
  const date = new Date(message.date * 1000);

  const month = date.toLocaleString('en-GB', {
    month: 'long',
    year: 'numeric',
    timeZone,
  });

  const day = date.toLocaleString('en-GB', {
    day: '2-digit',
    weekday: 'long',
    timeZone,
  });

  const time = date.toLocaleString('en-GB', {
    hour: '2-digit',
    minute: 'numeric',
    hour12: false,
    timeZone,
  });

  return { month, day, time, date };
};

const isLessThanXMinutesOff = (date1: Date, date2: Date, gap: number) => {
  console.log('less check', gap);
  if (gap === 0) return false;

  const diff = Math.abs(date1.getTime() - date2.getTime());
  console.log('diff', diff);
  const mins = Math.floor(diff / 1000 / 60);
  console.log('mins', mins);
  return Math.floor(diff / 1000 / 60) < gap;
};

export const makeExportFile = async (messages: IMessage[], chatId: number) => {
  if (!messages.length) return '';

  const result: string[] = [];

  let last: FormattedMessageDates | null = null;

  const settings = await getSettings(chatId);

  messages.forEach((msg) => {
    const dates = getDates(msg, settings.timezone);

    if (!last || last.month !== dates.month) {
      if (result.length) result.push('', '');
      result.push('### ' + dates.month);
    }

    if (!last || last.day !== dates.day) {
      result.push('##### ' + dates.day);
    }

    if (
      !last ||
      (last.time !== dates.time &&
        !(
          last &&
          last.month === dates.month &&
          last.day === dates.day &&
          isLessThanXMinutesOff(
            dates.date,
            last.date,
            settings.groupMessagesMinutes
          )
        ))
    ) {
      result.push(dates.time + ' ' + msg.message);
    } else {
      result.push(msg.message);
    }

    last = dates;
  });

  return result.join('\n');
};
