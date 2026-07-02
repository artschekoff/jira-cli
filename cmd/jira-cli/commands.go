package main

import "github.com/spf13/cobra"

func newSearchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "search",
		Short: "Search for issues (stub)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}

func newViewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "view",
		Short: "View an issue (stub)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}

func newCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create",
		Short: "Create an issue (stub)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}

func newEditCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "edit",
		Short: "Edit an issue (stub)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}

func newTransitionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "transition",
		Short: "Transition an issue (stub)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}

func newAssignCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "assign",
		Short: "Assign an issue (stub)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}

func newCommentCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "comment",
		Short: "Comment on an issue (stub)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}

func newSprintCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sprint",
		Short: "Sprint operations (stub)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}

func newProjectCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "project",
		Short: "Project operations (stub)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}

func newBoardCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "board",
		Short: "Board operations (stub)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}
