import { ChevronLeft, ChevronRight } from "lucide-react";
import { Button } from "@/components/ui/button";

export function CollectionPagination({ page, limit, total, onPageChange }: { page: number; limit: number; total: number; onPageChange: (page: number) => void }) {
  const pages = Math.max(1, Math.ceil(total / limit));
  return <div className="mt-6 flex flex-wrap items-center justify-between gap-3 border-t border-border pt-5 text-xs text-muted-foreground">
    <span>{total ? `${(page - 1) * limit + 1}–${Math.min(page * limit, total)} of ${total}` : "0 results"}</span>
    <nav aria-label="Collection pages" className="flex items-center gap-3">
      <span>Page {page} of {pages}</span>
      <Button variant="outline" size="icon" className="size-8" aria-label="Previous page" disabled={page <= 1} onClick={() => onPageChange(page - 1)}><ChevronLeft size={16}/></Button>
      <Button variant="outline" size="icon" className="size-8" aria-label="Next page" disabled={page >= pages} onClick={() => onPageChange(page + 1)}><ChevronRight size={16}/></Button>
    </nav>
  </div>;
}
