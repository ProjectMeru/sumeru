/** rAF render queue (OWL Fiber concept, SWC naming). */

type RenderJob = () => void;

const queue = new Set<RenderJob>();
let scheduled = false;

export function scheduleRender(job?: RenderJob): void {
  if (job) queue.add(job);
  if (scheduled) return;
  scheduled = true;
  requestAnimationFrame(flush);
}

function flush(): void {
  scheduled = false;
  const jobs = [...queue];
  queue.clear();
  for (const job of jobs) {
    try {
      job();
    } catch (err) {
      console.error("[SWC scheduler]", err);
    }
  }
}

export function cancelRender(job: RenderJob): void {
  queue.delete(job);
}
