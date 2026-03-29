import Convert from 'ansi-to-html';
import { ArrowDown } from 'lucide-react';
import { useEffect, useMemo, useRef, useState } from 'react';

import { Button } from '@/shared/components/generic/ui/button.component';
import { ScrollArea } from '@/shared/components/generic/ui/scroll-area.component';
import type { LogLine } from '@/shared/services/sse.service';

const converter = new Convert({ fg: '#d4d4d4', bg: '#09090b', newline: false });

interface LogViewerProps {
  logs: LogLine[];
  isStreaming: boolean;
}

export function LogViewer({ logs, isStreaming }: LogViewerProps) {
  const bottomRef = useRef<HTMLDivElement>(null);
  const sentinelRef = useRef<HTMLDivElement>(null);
  const [isAtBottom, setIsAtBottom] = useState(true);
  const [showTimestamps, setShowTimestamps] = useState(false);

  useEffect(() => {
    const sentinel = sentinelRef.current;
    if (!sentinel) return;

    const observer = new IntersectionObserver(
      ([entry]) => {
        setIsAtBottom(entry.isIntersecting);
      },
      { threshold: 0.1 }
    );

    observer.observe(sentinel);
    return () => observer.disconnect();
  }, []);

  useEffect(() => {
    if (isAtBottom) {
      bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
    }
  }, [logs.length, isAtBottom]);

  const scrollToBottom = () => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  const renderedLines = useMemo(
    () =>
      logs.map((line, i) => {
        if (line.separator) {
          return (
            <div key={i} className="flex items-center gap-3 px-4 py-3">
              <div className="h-px flex-1 bg-zinc-700" />
              <span className="shrink-0 text-xs font-medium text-zinc-400">{line.text}</span>
              <div className="h-px flex-1 bg-zinc-700" />
            </div>
          );
        }

        const html = converter.toHtml(line.text);
        const time = new Date(line.time).toLocaleTimeString();
        return (
          <div key={i} className="flex gap-3 px-4 py-0.5 hover:bg-white/5">
            {showTimestamps && (
              <span className="shrink-0 text-zinc-600 select-none">{time}</span>
            )}
            <span dangerouslySetInnerHTML={{ __html: html }} />
          </div>
        );
      }),
    [logs, showTimestamps]
  );

  return (
    <div className="relative">
      <div className="mb-2 flex justify-end">
        <label className="flex items-center gap-2 text-sm text-muted-foreground cursor-pointer select-none">
          <input
            type="checkbox"
            checked={showTimestamps}
            onChange={e => setShowTimestamps(e.target.checked)}
            className="rounded"
          />
          Show timestamps
        </label>
      </div>

      <ScrollArea className="h-[500px] overflow-auto rounded-lg bg-zinc-950 text-sm">
        <div className="py-2 font-mono">
          {renderedLines}
          {isStreaming && logs.length === 0 && (
            <div className="px-4 py-2 text-zinc-600">Waiting for output...</div>
          )}
          <div ref={sentinelRef} />
          <div ref={bottomRef} />
        </div>
      </ScrollArea>

      {!isAtBottom && (
        <Button
          variant="secondary"
          size="sm"
          className="absolute right-4 bottom-4"
          onClick={scrollToBottom}
        >
          <ArrowDown className="size-3.5" data-icon="inline-start" />
          Scroll to bottom
        </Button>
      )}
    </div>
  );
}
