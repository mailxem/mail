import { useState, useEffect } from "react";
import { useTeam } from "@/app/providers/team-provider";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { format } from "date-fns";
import { Loader2, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Input } from "@/components/ui/input";
import { WebhookEventType } from "@/lib";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { ScrollArea } from "@/components/ui/scroll-area";
import { toast } from "sonner";

interface WebhookDelivery {
  id: string;
  webhookId: string;
  eventType: WebhookEventType;
  status: number;
  payload: any;
  response?: string;
  error?: string;
  createdAt: string;
}

interface PaginationInfo {
  total: number;
  limit: number;
  offset: number;
  hasMore: boolean;
}

export function WebhookDeliveries({ webhookId }: { webhookId: string }) {
  const [deliveries, setDeliveries] = useState<WebhookDelivery[]>([]);
  const [loading, setLoading] = useState(true);
  const [pagination, setPagination] = useState<PaginationInfo>({
    total: 0,
    limit: 50,
    offset: 0,
    hasMore: false,
  });
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [selectedDelivery, setSelectedDelivery] =
    useState<WebhookDelivery | null>(null);

  const fetchDeliveries = async (offset = 0) => {
    try {
      setLoading(true);
      const params = new URLSearchParams({
        limit: pagination.limit.toString(),
        offset: offset.toString(),
        ...(statusFilter !== "all" && { status: statusFilter }),
      });

      const response = await fetch(
        `/api/webhooks/${webhookId}/deliveries?${params}`
      );
      if (!response.ok) throw new Error("Failed to fetch deliveries");

      const data = await response.json();
      setDeliveries(data.deliveries);
      setPagination(data.pagination);
    } catch (error) {
      toast.error("Failed to load webhook deliveries");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchDeliveries();
  }, [webhookId, statusFilter]);

  const getStatusBadgeVariant = (status: number) => {
    if (status >= 200 && status < 300) return "success";
    if (status >= 400 && status < 500) return "default";
    if (status >= 500) return "destructive";
    return "secondary";
  };

  const loadMore = () => {
    if (pagination.hasMore) {
      fetchDeliveries(pagination.offset + pagination.limit);
    }
  };

  return (
    <div className="space-y-4">
      <div className="flex justify-between items-center">
        <div className="flex items-center gap-2">
          <Select value={statusFilter} onValueChange={setStatusFilter}>
            <SelectTrigger className="w-[180px]">
              <SelectValue placeholder="Filter by status" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Status</SelectItem>
              <SelectItem value="200">Success (200)</SelectItem>
              <SelectItem value="400">Client Error (400)</SelectItem>
              <SelectItem value="500">Server Error (500)</SelectItem>
            </SelectContent>
          </Select>

          <Button
            variant="outline"
            size="icon"
            onClick={() => fetchDeliveries()}
            disabled={loading}
          >
            <RefreshCw className={`h-4 w-4 ${loading ? "animate-spin" : ""}`} />
          </Button>
        </div>

        <div className="text-sm text-muted-foreground">
          Showing {deliveries.length} of {pagination.total} deliveries
        </div>
      </div>

      <div className="border rounded-lg">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Event Type</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Time</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {deliveries.map((delivery) => (
              <TableRow key={delivery.id}>
                <TableCell>
                  <Badge variant="outline">{delivery.eventType}</Badge>
                </TableCell>
                <TableCell>
                  <Badge variant={getStatusBadgeVariant(delivery.status)}>
                    {delivery.status}
                  </Badge>
                </TableCell>
                <TableCell>
                  {format(new Date(delivery.createdAt), "PPp")}
                </TableCell>
                <TableCell className="text-right">
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => setSelectedDelivery(delivery)}
                  >
                    View Details
                  </Button>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>

      {pagination.hasMore && (
        <div className="flex justify-center">
          <Button variant="outline" onClick={loadMore} disabled={loading}>
            {loading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
            Load More
          </Button>
        </div>
      )}

      <Dialog
        open={!!selectedDelivery}
        onOpenChange={() => setSelectedDelivery(null)}
      >
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Webhook Delivery Details</DialogTitle>
            <DialogDescription>
              Delivery attempt at{" "}
              {selectedDelivery &&
                format(new Date(selectedDelivery.createdAt), "PPp")}
            </DialogDescription>
          </DialogHeader>

          {selectedDelivery && (
            <div className="space-y-4">
              <div>
                <h4 className="font-medium mb-2">Request Payload</h4>
                <ScrollArea className="h-[200px] rounded-md border p-4">
                  <pre className="text-sm">
                    {JSON.stringify(selectedDelivery.payload, null, 2)}
                  </pre>
                </ScrollArea>
              </div>

              <div>
                <h4 className="font-medium mb-2">Response</h4>
                <ScrollArea className="h-[200px] rounded-md border p-4">
                  <pre className="text-sm">
                    {selectedDelivery.response ||
                      selectedDelivery.error ||
                      "No response"}
                  </pre>
                </ScrollArea>
              </div>

              <div className="flex items-center gap-2">
                <span className="text-sm text-muted-foreground">Status:</span>
                <Badge variant={getStatusBadgeVariant(selectedDelivery.status)}>
                  {selectedDelivery.status}
                </Badge>
              </div>
            </div>
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}
