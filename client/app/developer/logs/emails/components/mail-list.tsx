"use client";

import { formatDistanceToNow } from "date-fns";

import { cn, generateKey } from "@/lib/utils";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Mail } from "@/app/types";
import { useMail } from "@/hooks/use-mail";
import { extract } from "letterparser";
import { parsedMailFrom, parsedMailTo } from "../utils";
import { IMAPEmail } from "@/types/imap";

interface MailListProps {
  items: Mail[] | IMAPEmail[];
  isFetchingNextPage: boolean;
  observerTarget: React.RefObject<HTMLDivElement>;
}

export default function MailList({
  items,
  isFetchingNextPage,
  observerTarget,
}: MailListProps) {
  const { mail, setMail } = useMail();

  const getMailText = (item: Mail | IMAPEmail) => {
    const { text } = extract(`Date: Wed, 01 Apr 2020 00:00:00 -0000
      From: A <a@example.com>
      To: B <b@example.com>
      Subject: Hello world!
      Mime-Version: 1.0
      Content-Type: text/html; charset=utf-8 
      
      ${item?.body}
      `);
    return text;
  };

  return (
    <ScrollArea className="h-[calc(100vh-200px)]">
      <div className="flex flex-col gap-2 pb-8 p-4 pt-0">
        {items?.map((item) => (
          <div
            key={generateKey()}
            className={cn(
              "flex flex-col cursor-pointer items-start gap-2 rounded-lg border border-muted p-3 text-left text-sm transition-all hover:bg-accent",
              mail?.id === item?.id && "bg-muted"
            )}
            onClick={() => {
              setMail({
                selected: item?.id,
                mail: item,
              });
            }}
          >
            <div className="flex w-full flex-col gap-1">
              <div className="flex items-center">
                <div className="flex items-center gap-2">
                  <div className="font-medium">{parsedMailTo(item)}</div>
                  {/* {!item.read && (
                    <span className="flex h-2 w-2 rounded-full bg-blue-600" />
                  )} */}
                </div>
                <div
                  className={cn(
                    "ml-auto text-xs",
                    mail?.id === item?.id
                      ? "text-foreground"
                      : "text-muted-foreground"
                  )}
                >
                  {item?.date || item?.createdAt
                    ? formatDistanceToNow(
                        new Date(item?.date ?? item?.createdAt),
                        {
                          addSuffix: true,
                        }
                      )
                    : formatDistanceToNow(new Date(), {
                        addSuffix: true,
                      })}
                </div>
              </div>
              <div className="text-xs font-medium">{item?.subject}</div>
            </div>
            <div className="line-clamp-2 text-xs text-muted-foreground">
              {getMailText(item).slice(0, 100)}
            </div>
            {/* {item.labels.length ? (
              <div className="flex items-center gap-2">
                {item?.labels?.map((label) => (
                  <Badge key={label} variant={getBadgeVariantFromLabel(label)}>
                    {label}
                  </Badge>
                ))}
              </div>
            ) : null} */}
          </div>
        ))}
        {!items ||
          (items?.length === 0 && (
            <div className="text-center p-4">No emails found</div>
          ))}
        <div ref={observerTarget} className="h-4" />
        {isFetchingNextPage && (
          <div className="text-center h-56">Loading more...</div>
        )}
      </div>
    </ScrollArea>
  );
}
