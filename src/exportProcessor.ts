import { IMessage } from './db/schema';

type FormattedMessageDates = {
  day: string;
  month: string;
  time: string;
};

const getDates = (message: IMessage): FormattedMessageDates => {
  const date = new Date(message.date * 1000);

  const month = date.toLocaleString('en-US', {
    month: 'long',
    year: '2-digit',
  });

  const day = date.toLocaleString('en-US', {
    day: '2-digit',
    weekday: 'long',
  });

  const time = date.toLocaleString('en-US', {
    hour: '2-digit',
    minute: 'numeric',
    hour12: false,
  });

  return { month, day, time };
};

export const makeExportFile = (messages: IMessage[]) => {
  if (!messages.length) return '';

  const result: string[] = [];

  let last: FormattedMessageDates | null = null;

  messages.forEach((msg) => {
    const dates = getDates(msg);

    if (!last || last.month !== dates.month) {
      if (result.length) result.push('');
      result.push('### ' + dates.month);
    }

    if (!last || last.day !== dates.day) {
      result.push('##### ' + dates.day);
    }

    if (!last || last.time !== dates.time) {
      result.push(dates.time + ' ' + msg.message);
    } else {
      result.push(msg.message);
    }

    last = dates;
  });

  return result.join('\n');
};
