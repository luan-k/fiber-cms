export function formatDate(isoString: string): string {
  const date = new Date(isoString);
  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  });
}

export function formatTime(isoString: string): string {
  const date = new Date(isoString);
  return date.toLocaleTimeString('en-US', {
    hour: '2-digit',
    minute: '2-digit',
    hour12: true,
  }).replace(/AM|PM/, (m) => m.toLowerCase());
}
export function formatDateTime(isoString: string): string {
  const date = new Date(isoString);
  // Format: Month Day, Year\nHH:mm am/pm
  const options: Intl.DateTimeFormatOptions = {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    hour12: true,
  };
  // e.g. "July 27, 2025, 09:28 PM"
  const formatted = date.toLocaleString('en-US', options);
  // Split date and time, and add a line break
  const [datePart, timePart, ...rest] = formatted.split(', ');
  // If timePart is undefined, fallback to formatted
  if (!timePart) return formatted;
  // timePart is like "09:28 PM", convert to lowercase am/pm
  const time = timePart.replace(/AM|PM/, (m) => m.toLowerCase());
  return `${datePart}, ${date.getFullYear()}\n${time}`;
}
export function formatFullDate(isoString: string): string {
  const date = new Date(isoString);
  return date.toLocaleString(undefined, {
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}