import type { SavedDiscoveryPosition } from "./discovery-position";
type Value = Omit<SavedDiscoveryPosition, "revision">;
type Status = "saving" | "saved" | "error" | "conflict" | "account-changed";
interface Transport {
  read: () => Promise<number>;
  write: (
    value: Value & { expectedRevision: number; expectedAccountId: string },
  ) => Promise<number>;
  status: (status: Status) => void;
  conflict: (error: unknown) => boolean;
  accountChanged: (error: unknown) => boolean;
}
const identity = (value: Value) => JSON.stringify(value);

/** Per-mounted-account queue: coalesce navigation and serialize revision writes. */
export class DiscoveryPositionWriter {
  private accountId: string;
  private accountChanged = false;
  private revision: number | null;
  private transport: Transport;
  private selected: Value | null = null;
  private confirmed = "";
  private timer: ReturnType<typeof setTimeout> | undefined;
  private running = false;
  private disposed = false;
  private blocked = false;
  constructor(
    accountId: string,
    revision: number | null,
    transport: Transport,
  ) {
    this.accountId = accountId;
    this.revision = revision;
    this.transport = transport;
  }
  select(value: Value) {
    if (this.disposed) return;
    this.selected = value;
    clearTimeout(this.timer);
    // Failed or stale writes remain uncertain until an explicit retry. An old
    // confirmation cannot hide that uncertainty or the control that resolves it.
    if (this.blocked) return;
    if (!this.running && this.confirmed === identity(value)) {
      this.transport.status("saved");
      return;
    }
    this.transport.status("saving");
    this.timer = setTimeout(() => void this.flush(), 180);
  }
  async flush() {
    clearTimeout(this.timer);
    if (this.disposed || this.running || this.blocked || !this.selected) return;
    this.running = true;
    try {
      if (this.revision === null) this.revision = await this.transport.read();
      while (
        !this.disposed &&
        this.selected &&
        identity(this.selected) !== this.confirmed
      ) {
        const value = this.selected;
        this.revision = await this.transport.write({
          ...value,
          expectedRevision: this.revision!,
          expectedAccountId: this.accountId,
        });
        if (this.disposed) return;
        this.confirmed = identity(value);
      }
      if (!this.disposed) this.transport.status("saved");
    } catch (error) {
      if (this.disposed) return;
      this.blocked = true;
      this.accountChanged = this.transport.accountChanged(error);
      this.transport.status(
        this.accountChanged
          ? "account-changed"
          : this.transport.conflict(error)
            ? "conflict"
            : "error",
      );
    } finally {
      this.running = false;
    }
  }
  async retry() {
    if (this.disposed || this.running || this.accountChanged) return;
    this.blocked = false;
    this.revision = null;
    // The failed write may have committed. After re-reading its revision, write
    // the selected destination again rather than trusting an older confirmation.
    this.confirmed = "";
    this.transport.status("saving");
    await this.flush();
  }
  dispose() {
    this.disposed = true;
    clearTimeout(this.timer);
  }
}
