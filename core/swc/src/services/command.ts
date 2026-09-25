export interface CommandDef {
  id: string;
  label: string;
  run: () => void;
  when?: () => boolean;
}

export class CommandService {
  private readonly commands = new Map<string, CommandDef>();

  register(def: CommandDef): void {
    if (!def.id.trim() || this.commands.has(def.id)) return;
    this.commands.set(def.id, def);
  }

  list(): CommandDef[] {
    return [...this.commands.values()].filter((cmd) => {
      if (!cmd.when) return true;
      try {
        return cmd.when();
      } catch {
        return false;
      }
    });
  }

  run(id: string): boolean {
    const cmd = this.commands.get(id);
    if (!cmd) return false;
    if (cmd.when) {
      try {
        if (!cmd.when()) return false;
      } catch {
        return false;
      }
    }
    cmd.run();
    return true;
  }
}
