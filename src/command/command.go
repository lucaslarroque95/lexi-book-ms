// Package command implements a reversible unit-of-work pattern for
// operations that touch more than one system (Postgres, object storage, a
// message queue) with no real distributed transaction to lean on. Each
// Command knows how to execute itself and how to undo itself; Run executes a
// sequence of commands and, if any of them fails, undoes every command that
// already succeeded, in reverse order.
package command

import "context"

type Command interface {
	Execute(ctx context.Context) error
	Undo(ctx context.Context)
}

func Run(ctx context.Context, commands ...Command) error {
	executed := make([]Command, 0, len(commands))
	for _, cmd := range commands {
		if err := cmd.Execute(ctx); err != nil {
			for i := len(executed) - 1; i >= 0; i-- {
				executed[i].Undo(ctx)
			}
			return err
		}
		executed = append(executed, cmd)
	}
	return nil
}
