import type { SavedDiscoveryPosition } from "./discovery-position";
type Value = Omit<SavedDiscoveryPosition, "revision">;
type Status = "saving" | "saved" | "error" | "conflict";
interface Transport {
  read: () => Promise<number>;
  write: (value: Value & { expectedRevision: number }) => Promise<number>;
  status: (status: Status) => void;
  conflict: (error: unknown) => boolean;
}
const identity = (value: Value) => JSON.stringify(value);

/** Per-mounted-account queue: coalesce navigation and serialize revision writes. */
export class DiscoveryPositionWriter {
  private revision: number | null;
  private transport: Transport;
  private selected: Value | null = null;
  private confirmed = "";
  private timer: ReturnType<typeof setTimeout> | undefined;
  private running = false;
  private disposed = false;
  private blocked = false;
  constructor(revision: number | null, transport: Transport) {
    this.revision = revision;
    this.transport = transport;
  }
  select(value: Value) {
    if (this.disposed) return;
    this.selected = value;
    clearTimeout(this.timer);
    if (this.confirmed === identity(value)) {
      this.transport.status("saved");
      return;
    }
    if (this.blocked) return;
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
        });
        if (this.disposed) return;
        this.confirmed = identity(value);
      }
      if (!this.disposed) this.transport.status("saved");
    } catch (error) {
      if (this.disposed) return;
      this.blocked = true;
      this.transport.status(
        this.transport.conflict(error) ? "conflict" : "error",
      );
    } finally {
      this.running = false;
    }
  }
  async retry() {
    if (this.disposed || this.running) return;
    this.blocked = false;
    this.revision = null;
    this.transport.status("saving");
    await this.flush();
  }
  dispose() {
    this.disposed = true;
    clearTimeout(this.timer);
  }
}
