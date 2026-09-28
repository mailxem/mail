import { addDays } from "date-fns";
import { addHours } from "date-fns";
import { format } from "date-fns";
import { nextSaturday } from "date-fns";
import {
  Archive,
  ArchiveX,
  Clock,
  Code,
  Copy,
  Forward,
  Loader2,
  MoreVertical,
  Reply,
  ReplyAll,
  Send,
  Trash2,
} from "lucide-react";

import {
  DropdownMenuContent,
  DropdownMenuItem,
} from "@/components/ui/dropdown-menu";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Calendar } from "@/components/ui/calendar";
import {
  DropdownMenu,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Label } from "@/components/ui/label";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { Separator } from "@/components/ui/separator";
import { Switch } from "@/components/ui/switch";
import { Textarea } from "@/components/ui/textarea";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { Mail } from "@/app/types";
import { useQuery } from "@tanstack/react-query";
import { extract } from "letterparser";
import { Letter } from "react-letter";
import { Fragment, useMemo, useState } from "react";
import { toast } from "sonner";
import { parsedMailFrom, parsedMailTo } from "../utils";
import { useApi } from "@/hooks/use-api";
import { HamburgerMenuIcon } from "@radix-ui/react-icons";
import {
  Sheet,
  SheetHeader,
  SheetContent,
  SheetTrigger,
  SheetTitle,
  SheetDescription,
} from "@/components/ui/sheet";
import JsonView, { ValueQuote } from "@uiw/react-json-view";
import { IMAPEmail } from "@/types/imap";

interface MailDisplayProps {
  mail: IMAPEmail | Mail | null;
  fetchMail?: boolean;
}

export default function MailDisplay({
  mail,
  fetchMail = true,
}: MailDisplayProps) {
  const today = new Date();
  const [reply, setReply] = useState("");
  const [viewSourceOpen, setViewSourceOpen] = useState(false);
  const { apiFetch } = useApi();

  const getMail = async (id: string) => {
    const response = await apiFetch(`/emails?id=${id}&include=Contact`);
    const data = await response.json();

    if (id) {
      const email = data.data?.[0];
      data.data[0].body = Buffer.from(email.body, "base64").toString("utf-8");
    }

    return data.data?.[0];
  };

  const getParsedMail = (email: { body: string }) => {
    const decodedBody = email?.body;
    const { html, text } = extract(`Date: Wed, 01 Apr 2020 00:00:00 -0000
From: A <a@example.com>
To: B <b@example.com>
Subject: Hello world!
Mime-Version: 1.0
Content-Type: text/html; charset=utf-8 

${decodedBody}
`);

    return { html, text };
  };

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    let email = extractEmail(mailToRender?.replyTo ?? mailToRender?.from);
    const response = await apiFetch(`/emails`, {
      method: "POST",
      body: JSON.stringify({
        to: email,
        html: `
        ${reply}
        <div style="padding: 20px; border-left: 2px solid #ccc; margin: 10px 0;">
          <div style="color: #666; margin-bottom: 10px;">On ${format(new Date(mailToRender?.createdAt ?? mailToRender?.date), "PPpp")}, ${mailToRender?.from} wrote:</div>
          ${getParsedMail(mailToRender)}
        </div>
        `,
        cc: mailToRender?.cc
          ?.split(",")
          .filter((cc) => cc !== email && cc !== "N/A"),
        bcc: mailToRender?.bcc
          ?.split(",")
          .filter((bcc) => bcc !== email && bcc !== "N/A"),
        subject: mailToRender?.subject ?? "Reply to email",
      }),
    });
    if (!response.ok) {
      toast.error("Failed to send email");
      return;
    }
    toast.success("Email sent successfully");
    setReply("");
  };

  const { data: fetchedMail, isLoading: isLoadingMail } = useQuery({
    queryKey: ["mail", mail?.id],
    queryFn: () => getMail(mail?.id),
    enabled: !!mail?.id && fetchMail,
  });

  const extractEmail = (email: string) => {
    const emailRegex = /<([^>]+)>/;
    const match = email.match(emailRegex);
    return match ? match[1] : email;
  };

  const mailToRender = useMemo(() => {
    if (fetchMail) {
      return fetchedMail;
    }
    return mail;
  }, [mail, fetchedMail, fetchMail]);

  const handleCopyEmail = async () => {
    if (!mailToRender) return;

    try {
      const emailContent = `From: ${mailToRender.from}
To: ${mailToRender.to}
Subject: ${mailToRender.subject}
Date: ${format(new Date(mailToRender.createdAt ?? mailToRender.date), "PPpp")}
${mailToRender.cc ? `CC: ${mailToRender.cc}` : ""}
${mailToRender.bcc ? `BCC: ${mailToRender.bcc}` : ""}
${mailToRender.replyTo ? `Reply-To: ${mailToRender.replyTo}` : ""}

${getParsedMail(mailToRender).text || getParsedMail(mailToRender).html}`;

      await navigator.clipboard.writeText(emailContent);
      toast.success("Email copied to clipboard");
    } catch (error) {
      toast.error("Failed to copy email");
    }
  };

  return (
    <div className="flex h-full flex-col">
      {!fetchMail && (
        <>
          <div className="flex items-center p-2">
            <div className="flex items-center gap-2">
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button variant="ghost" size="icon" disabled={!fetchedMail}>
                    <Archive className="h-4 w-4" />
                    <span className="sr-only">Archive</span>
                  </Button>
                </TooltipTrigger>
                <TooltipContent>Archive</TooltipContent>
              </Tooltip>
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button variant="ghost" size="icon" disabled={!fetchedMail}>
                    <ArchiveX className="h-4 w-4" />
                    <span className="sr-only">Move to junk</span>
                  </Button>
                </TooltipTrigger>
                <TooltipContent>Move to junk</TooltipContent>
              </Tooltip>
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button variant="ghost" size="icon" disabled={!fetchedMail}>
                    <Trash2 className="h-4 w-4" />
                    <span className="sr-only">Move to trash</span>
                  </Button>
                </TooltipTrigger>
                <TooltipContent>Move to trash</TooltipContent>
              </Tooltip>
              <Separator orientation="vertical" className="mx-1 h-6" />
              <Tooltip>
                <Popover>
                  <PopoverTrigger asChild>
                    <TooltipTrigger asChild>
                      <Button
                        variant="ghost"
                        size="icon"
                        disabled={!fetchedMail}
                      >
                        <Clock className="h-4 w-4" />
                        <span className="sr-only">Snooze</span>
                      </Button>
                    </TooltipTrigger>
                  </PopoverTrigger>
                  <PopoverContent className="flex w-[535px] p-0">
                    <div className="flex flex-col gap-2 border-r px-2 py-4">
                      <div className="px-4 text-sm font-medium">
                        Snooze until
                      </div>
                      <div className="grid min-w-[250px] gap-1">
                        <Button
                          variant="ghost"
                          className="justify-start font-normal"
                        >
                          Later today{" "}
                          <span className="ml-auto text-muted-foreground">
                            {format(addHours(today, 4), "E, h:m b")}
                          </span>
                        </Button>
                        <Button
                          variant="ghost"
                          className="justify-start font-normal"
                        >
                          Tomorrow
                          <span className="ml-auto text-muted-foreground">
                            {format(addDays(today, 1), "E, h:m b")}
                          </span>
                        </Button>
                        <Button
                          variant="ghost"
                          className="justify-start font-normal"
                        >
                          This weekend
                          <span className="ml-auto text-muted-foreground">
                            {format(nextSaturday(today), "E, h:m b")}
                          </span>
                        </Button>
                        <Button
                          variant="ghost"
                          className="justify-start font-normal"
                        >
                          Next week
                          <span className="ml-auto text-muted-foreground">
                            {format(addDays(today, 7), "E, h:m b")}
                          </span>
                        </Button>
                      </div>
                    </div>
                    <div className="p-2">
                      <Calendar />
                    </div>
                  </PopoverContent>
                </Popover>
                <TooltipContent>Snooze</TooltipContent>
              </Tooltip>
            </div>
            <div className="ml-auto flex items-center gap-2">
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button variant="ghost" size="icon" disabled={!mailToRender}>
                    <Reply className="h-4 w-4" />
                    <span className="sr-only">Reply</span>
                  </Button>
                </TooltipTrigger>
                <TooltipContent>Reply</TooltipContent>
              </Tooltip>
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button variant="ghost" size="icon" disabled={!mailToRender}>
                    <ReplyAll className="h-4 w-4" />
                    <span className="sr-only">Reply all</span>
                  </Button>
                </TooltipTrigger>
                <TooltipContent>Reply all</TooltipContent>
              </Tooltip>
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button variant="ghost" size="icon" disabled={!mailToRender}>
                    <Forward className="h-4 w-4" />
                    <span className="sr-only">Forward</span>
                  </Button>
                </TooltipTrigger>
                <TooltipContent>Forward</TooltipContent>
              </Tooltip>
            </div>
            <Separator orientation="vertical" className="mx-2 h-6" />
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="icon" disabled={!mailToRender}>
                  <MoreVertical className="h-4 w-4" />
                  <span className="sr-only">More</span>
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem>Mark as unread</DropdownMenuItem>
                <DropdownMenuItem>Star thread</DropdownMenuItem>
                <DropdownMenuItem>Add label</DropdownMenuItem>
                <DropdownMenuItem>Mute thread</DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
          <Separator />
        </>
      )}
      {mail && !isLoadingMail ? (
        <div className="flex flex-1 flex-col">
          <div className="flex items-start p-4">
            <div className="flex items-start gap-4 text-sm">
              <Avatar className="bg-red-500 dark:bg-white">
                <AvatarImage
                  src={`https://api.dicebear.com/9.x/lorelei/svg?seed=${parsedMailTo(mailToRender)}`}
                  alt={parsedMailTo(mailToRender)?.toString()}
                />
                <AvatarFallback>
                  {parsedMailTo(mailToRender)
                    ?.toString()
                    .split(" ")[0]
                    ?.toUpperCase()}
                </AvatarFallback>
              </Avatar>
              <div className="grid gap-1">
                <div className="font-medium">
                  {parsedMailTo(mailToRender)}
                </div>
                <div className="line-clamp-1 text-xs">
                  Sent from: {parsedMailFrom(mailToRender)}
                </div>
                <div className="line-clamp-1 text-xs">
                  {mailToRender?.subject}
                </div>
                {mailToRender?.replyTo && (
                  <div className="line-clamp-1 text-xs">
                    <span className="font-medium">Reply-To:</span>{" "}
                    {mailToRender?.replyTo}
                  </div>
                )}
                {mailToRender?.cc && (
                  <div className="line-clamp-1 text-xs">
                    <span className="font-medium">CC:</span> {mailToRender?.cc}
                  </div>
                )}
                {mailToRender?.bcc && (
                  <div className="line-clamp-1 text-xs">
                    <span className="font-medium">BCC:</span>{" "}
                    {mailToRender?.bcc}
                  </div>
                )}
              </div>
            </div>
            <div className="flex flex-col items-end ml-auto gap-2">
              {(mailToRender?.createdAt || mailToRender?.date) && (
                <div className="ml-auto text-xs text-muted-foreground">
                  {format(
                    new Date(mailToRender.createdAt ?? mailToRender.date),
                    "PPpp"
                  )}
                </div>
              )}
              <Popover>
                <PopoverTrigger asChild>
                  <Button variant="outline" size="icon">
                    <HamburgerMenuIcon className="size-4" />
                  </Button>
                </PopoverTrigger>
                <PopoverContent className="flex flex-col gap-2 w-max">
                  <Sheet>
                    <SheetTrigger asChild>
                      <Button
                        variant="ghost"
                        className="w-full items-start justify-start"
                      >
                        <Code className="h-4 w-4" />
                        <span>See Email Variables</span>
                      </Button>
                    </SheetTrigger>
                    <SheetContent className="sm:max-w-screen-xl max-h-screen overflow-y-auto">
                      <SheetHeader className="mb-4">
                        <SheetTitle>Email Variables</SheetTitle>
                      </SheetHeader>
                      <JsonView value={mailToRender?.data ?? {}}>
                        <JsonView.Url
                          render={(props, { type, value }) => {
                            if (type === "type" && value instanceof URL) {
                              return <span />;
                            }
                            if (type === "value" && value instanceof URL) {
                              return (
                                <Fragment>
                                  <a
                                    href={value.href}
                                    target="_blank"
                                    {...props}
                                  >
                                    <ValueQuote />
                                    {value.href}
                                    <ValueQuote />
                                  </a>
                                  Open URL
                                </Fragment>
                              );
                            }
                          }}
                        />
                      </JsonView>
                    </SheetContent>
                  </Sheet>
                  <Button
                    variant="ghost"
                    className="w-full items-start justify-start"
                    onClick={handleCopyEmail}
                  >
                    <Copy className="h-4 w-4" />
                    <span>Copy Email</span>
                  </Button>
                  <Sheet open={viewSourceOpen} onOpenChange={setViewSourceOpen}>
                    <SheetTrigger asChild>
                      <Button
                        variant="ghost"
                        className="w-full items-start justify-start"
                      >
                        <Code className="h-4 w-4" />
                        <span>View Source</span>
                      </Button>
                    </SheetTrigger>
                    <SheetContent className="sm:max-w-screen-xl max-h-screen overflow-hidden flex flex-col">
                      <SheetHeader className="mb-4">
                        <SheetTitle>Email Source</SheetTitle>
                        <SheetDescription>
                          Raw email content and headers
                        </SheetDescription>
                      </SheetHeader>
                      <div className="flex-1 overflow-hidden">
                        <pre className="h-full overflow-auto text-xs bg-muted p-4 rounded-lg">
                          <code>{mailToRender?.body}</code>
                        </pre>
                      </div>
                    </SheetContent>
                  </Sheet>
                </PopoverContent>
              </Popover>
            </div>
          </div>
          <Separator />
          <Letter html={getParsedMail(mailToRender).html} />
          <Separator className="mt-auto" />
          {
            <div className="py-4 ml-1 mr-4">
              <form onSubmit={handleSubmit}>
                <div className="grid gap-4">
                  <Textarea
                    className="p-2 border-muted"
                    name="reply"
                    onChange={(e) => {
                      setReply(e.target.value);
                    }}
                    placeholder={`${fetchedMail ? "Follow up with" : "Reply to"} ${parsedMailTo(mailToRender)}...`}
                    disabled={isLoadingMail}
                  />
                  {
                    <div className="flex gap-8 items-center">
                      <Button type="submit" className="mr-auto px-8">
                        <Send className="h-4 w-4" />
                        Send
                      </Button>
                      {!fetchMail && (
                        <Label
                          htmlFor="mute"
                          className="flex items-center gap-2 text-xs font-normal"
                        >
                          <Switch id="mute" aria-label="Mute thread" /> Mute
                          this thread
                        </Label>
                      )}
                    </div>
                  }
                </div>
              </form>
            </div>
          }
        </div>
      ) : (
        <div className="p-8 text-center text-muted-foreground">
          {isLoadingMail ? (
            <div className="flex py-32 items-center justify-center">
              <Loader2 className="h-4 w-4 animate-spin" />
            </div>
          ) : (
            "No message selected"
          )}
        </div>
      )}
    </div>
  );
}
