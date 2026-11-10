import Image from "next/image";
import {
  Archive,
  ArrowLeft,
  Mail,
  MoreHorizontal,
  Reply,
  Star,
  Trash2,
} from "lucide-react";

// A Xem newsletter template open in a phone's mail app. The email scrolls
// slowly by itself; reduced motion keeps it still at the top.
export function PhonePreview() {
  return (
    <div className="pointer-events-none mx-auto w-[clamp(190px,34%,300px)] rounded-[2.3rem] border border-black/80 bg-[#111] p-[7px] shadow-[0_30px_80px_-20px_rgba(0,0,0,.55)]">
      <style>{`@keyframes phone-mail-scroll{0%,12%{transform:translateY(0)}88%,100%{transform:translateY(calc(-100% + 100cqh))}}`}</style>
      <div className="flex aspect-[9/19] flex-col overflow-hidden rounded-[1.9rem] bg-white text-[#1f1f1f]">
        <div className="relative flex items-center justify-between px-5 pb-1 pt-2.5 text-[10px] font-semibold">
          <span>9:41</span>
          <span className="absolute left-1/2 top-1.5 h-[18px] w-[34%] -translate-x-1/2 rounded-full bg-black" />
          <span className="flex items-center gap-1">
            <span className="flex items-end gap-px">
              {[3, 5, 7, 9].map((h) => (
                <i
                  key={h}
                  className="block w-[2px] rounded-sm bg-current"
                  style={{ height: h }}
                />
              ))}
            </span>
            <span className="ml-1 h-[9px] w-[18px] rounded-[3px] border border-current p-px">
              <span className="block h-full w-3/4 rounded-[1px] bg-[#34c759]" />
            </span>
          </span>
        </div>
        <div className="flex items-center justify-between px-3 py-2 text-[#444]">
          <ArrowLeft size={14} />
          <span className="flex gap-3">
            <Archive size={13} />
            <Trash2 size={13} />
            <Mail size={13} />
            <MoreHorizontal size={13} />
          </span>
        </div>
        <div className="flex items-start justify-between gap-2 px-3.5">
          <p className="text-[13px] leading-snug">
            Reading’s good for you{" "}
            <span className="ml-0.5 rounded bg-black/5 px-1 py-px align-middle text-[8px] text-[#555]">
              Inbox
            </span>
          </p>
          <Star size={13} className="mt-0.5 shrink-0 text-[#666]" />
        </div>
        <div className="flex items-center gap-2 px-3.5 pb-2 pt-2.5">
          <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-[#dfd2c2] text-[10px] font-semibold">
            S
          </span>
          <span className="min-w-0 flex-1 leading-tight">
            <span className="block truncate text-[10.5px] font-semibold">
              The Sunday Edit{" "}
              <span className="font-normal text-[#777]">9:41 AM</span>
            </span>
            <span className="block text-[9px] text-[#777]">to me</span>
          </span>
          <Reply size={12} className="text-[#666]" />
        </div>
        <div className="relative flex-1 overflow-hidden border-t border-black/5 [container-type:size]">
          <Image
            src="/images/templates/reading-good-for-you-full.webp"
            alt="The “Reading’s Good for You” Xem newsletter template, shown in a phone’s mail app"
            width={640}
            height={3141}
            sizes="300px"
            className="block w-full animate-[phone-mail-scroll_18s_ease-in-out_infinite_alternate] motion-reduce:animate-none"
          />
        </div>
      </div>
    </div>
  );
}
