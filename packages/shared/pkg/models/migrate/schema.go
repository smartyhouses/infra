// Code generated by ent, DO NOT EDIT.

package migrate

import (
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/dialect/sql/schema"
	"entgo.io/ent/schema/field"
)

var (
	// AccessTokensColumns holds the columns for the "access_tokens" table.
	AccessTokensColumns = []*schema.Column{
		{Name: "access_token", Type: field.TypeString, Unique: true, SchemaType: map[string]string{"postgres": "text"}},
		{Name: "created_at", Type: field.TypeTime, Default: "CURRENT_TIMESTAMP"},
		{Name: "user_id", Type: field.TypeUUID},
	}
	// AccessTokensTable holds the schema information for the "access_tokens" table.
	AccessTokensTable = &schema.Table{
		Name:       "access_tokens",
		Columns:    AccessTokensColumns,
		PrimaryKey: []*schema.Column{AccessTokensColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "access_tokens_users_access_tokens",
				Columns:    []*schema.Column{AccessTokensColumns[2]},
				RefColumns: []*schema.Column{UsersColumns[0]},
				OnDelete:   schema.Cascade,
			},
		},
	}
	// EnvsColumns holds the columns for the "envs" table.
	EnvsColumns = []*schema.Column{
		{Name: "id", Type: field.TypeString, Unique: true, SchemaType: map[string]string{"postgres": "text"}},
		{Name: "created_at", Type: field.TypeTime, Default: "CURRENT_TIMESTAMP"},
		{Name: "updated_at", Type: field.TypeTime},
		{Name: "public", Type: field.TypeBool, Default: "false"},
		{Name: "build_count", Type: field.TypeInt32, Default: 1},
		{Name: "spawn_count", Type: field.TypeInt64, Comment: "Number of times the env was spawned", Default: 0},
		{Name: "last_spawned_at", Type: field.TypeTime, Nullable: true, Comment: "Timestamp of the last time the env was spawned"},
		{Name: "team_id", Type: field.TypeUUID},
	}
	// EnvsTable holds the schema information for the "envs" table.
	EnvsTable = &schema.Table{
		Name:       "envs",
		Columns:    EnvsColumns,
		PrimaryKey: []*schema.Column{EnvsColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "envs_teams_envs",
				Columns:    []*schema.Column{EnvsColumns[7]},
				RefColumns: []*schema.Column{TeamsColumns[0]},
				OnDelete:   schema.NoAction,
			},
		},
	}
	// EnvAliasesColumns holds the columns for the "env_aliases" table.
	EnvAliasesColumns = []*schema.Column{
		{Name: "alias", Type: field.TypeString, Unique: true, SchemaType: map[string]string{"postgres": "text"}},
		{Name: "is_renamable", Type: field.TypeBool, Default: true},
		{Name: "env_id", Type: field.TypeString, SchemaType: map[string]string{"postgres": "text"}},
	}
	// EnvAliasesTable holds the schema information for the "env_aliases" table.
	EnvAliasesTable = &schema.Table{
		Name:       "env_aliases",
		Columns:    EnvAliasesColumns,
		PrimaryKey: []*schema.Column{EnvAliasesColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "env_aliases_envs_env_aliases",
				Columns:    []*schema.Column{EnvAliasesColumns[2]},
				RefColumns: []*schema.Column{EnvsColumns[0]},
				OnDelete:   schema.Cascade,
			},
		},
	}
	// EnvBuildsColumns holds the columns for the "env_builds" table.
	EnvBuildsColumns = []*schema.Column{
		{Name: "id", Type: field.TypeUUID, Unique: true, Default: "gen_random_uuid()"},
		{Name: "created_at", Type: field.TypeTime, Default: "CURRENT_TIMESTAMP"},
		{Name: "updated_at", Type: field.TypeTime},
		{Name: "finished_at", Type: field.TypeTime, Nullable: true},
		{Name: "status", Type: field.TypeEnum, Enums: []string{"waiting", "building", "failed", "success"}, Default: "waiting", SchemaType: map[string]string{"postgres": "text"}},
		{Name: "dockerfile", Type: field.TypeString, Nullable: true, SchemaType: map[string]string{"postgres": "text"}},
		{Name: "start_cmd", Type: field.TypeString, Nullable: true, SchemaType: map[string]string{"postgres": "text"}},
		{Name: "vcpu", Type: field.TypeInt64},
		{Name: "ram_mb", Type: field.TypeInt64},
		{Name: "free_disk_size_mb", Type: field.TypeInt64},
		{Name: "total_disk_size_mb", Type: field.TypeInt64, Nullable: true},
		{Name: "kernel_version", Type: field.TypeString, Default: "vmlinux-5.10.186", SchemaType: map[string]string{"postgres": "text"}},
		{Name: "firecracker_version", Type: field.TypeString, Default: "v1.7.0-dev_8bb88311", SchemaType: map[string]string{"postgres": "text"}},
		{Name: "envd_version", Type: field.TypeString, Nullable: true, SchemaType: map[string]string{"postgres": "text"}},
		{Name: "env_id", Type: field.TypeString, Nullable: true, SchemaType: map[string]string{"postgres": "text"}},
	}
	// EnvBuildsTable holds the schema information for the "env_builds" table.
	EnvBuildsTable = &schema.Table{
		Name:       "env_builds",
		Columns:    EnvBuildsColumns,
		PrimaryKey: []*schema.Column{EnvBuildsColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "env_builds_envs_builds",
				Columns:    []*schema.Column{EnvBuildsColumns[14]},
				RefColumns: []*schema.Column{EnvsColumns[0]},
				OnDelete:   schema.Cascade,
			},
		},
	}
	// TeamsColumns holds the columns for the "teams" table.
	TeamsColumns = []*schema.Column{
		{Name: "id", Type: field.TypeUUID, Unique: true, Default: "gen_random_uuid()"},
		{Name: "created_at", Type: field.TypeTime, Default: "CURRENT_TIMESTAMP"},
		{Name: "is_default", Type: field.TypeBool},
		{Name: "is_banned", Type: field.TypeBool, Default: "false"},
		{Name: "is_blocked", Type: field.TypeBool, Default: "false"},
		{Name: "blocked_reason", Type: field.TypeString, Nullable: true, SchemaType: map[string]string{"postgres": "text"}},
		{Name: "name", Type: field.TypeString, SchemaType: map[string]string{"postgres": "text"}},
		{Name: "email", Type: field.TypeString, Size: 255, SchemaType: map[string]string{"postgres": "character varying(255)"}},
		{Name: "tier", Type: field.TypeString, SchemaType: map[string]string{"postgres": "text"}},
	}
	// TeamsTable holds the schema information for the "teams" table.
	TeamsTable = &schema.Table{
		Name:       "teams",
		Columns:    TeamsColumns,
		PrimaryKey: []*schema.Column{TeamsColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "teams_tiers_teams",
				Columns:    []*schema.Column{TeamsColumns[8]},
				RefColumns: []*schema.Column{TiersColumns[0]},
				OnDelete:   schema.NoAction,
			},
		},
	}
	// TeamAPIKeysColumns holds the columns for the "team_api_keys" table.
	TeamAPIKeysColumns = []*schema.Column{
		{Name: "api_key", Type: field.TypeString, Unique: true, SchemaType: map[string]string{"postgres": "character varying(44)"}},
		{Name: "created_at", Type: field.TypeTime, Default: "CURRENT_TIMESTAMP"},
		{Name: "team_id", Type: field.TypeUUID},
	}
	// TeamAPIKeysTable holds the schema information for the "team_api_keys" table.
	TeamAPIKeysTable = &schema.Table{
		Name:       "team_api_keys",
		Columns:    TeamAPIKeysColumns,
		PrimaryKey: []*schema.Column{TeamAPIKeysColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "team_api_keys_teams_team_api_keys",
				Columns:    []*schema.Column{TeamAPIKeysColumns[2]},
				RefColumns: []*schema.Column{TeamsColumns[0]},
				OnDelete:   schema.Cascade,
			},
		},
	}
	// TiersColumns holds the columns for the "tiers" table.
	TiersColumns = []*schema.Column{
		{Name: "id", Type: field.TypeString, Unique: true, SchemaType: map[string]string{"postgres": "text"}},
		{Name: "name", Type: field.TypeString, SchemaType: map[string]string{"postgres": "text"}},
		{Name: "disk_mb", Type: field.TypeInt64, Default: "512"},
		{Name: "concurrent_instances", Type: field.TypeInt64, Comment: "The number of instances the team can run concurrently"},
		{Name: "max_length_hours", Type: field.TypeInt64},
	}
	// TiersTable holds the schema information for the "tiers" table.
	TiersTable = &schema.Table{
		Name:       "tiers",
		Columns:    TiersColumns,
		PrimaryKey: []*schema.Column{TiersColumns[0]},
	}
	// UsersColumns holds the columns for the "users" table.
	UsersColumns = []*schema.Column{
		{Name: "id", Type: field.TypeUUID, Unique: true, Default: "gen_random_uuid()"},
		{Name: "email", Type: field.TypeString, Size: 255, SchemaType: map[string]string{"postgres": "character varying(255)"}},
	}
	// UsersTable holds the schema information for the "users" table.
	UsersTable = &schema.Table{
		Name:       "users",
		Columns:    UsersColumns,
		PrimaryKey: []*schema.Column{UsersColumns[0]},
	}
	// UsersTeamsColumns holds the columns for the "users_teams" table.
	UsersTeamsColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "is_default", Type: field.TypeBool, Default: false},
		{Name: "user_id", Type: field.TypeUUID},
		{Name: "team_id", Type: field.TypeUUID},
	}
	// UsersTeamsTable holds the schema information for the "users_teams" table.
	UsersTeamsTable = &schema.Table{
		Name:       "users_teams",
		Columns:    UsersTeamsColumns,
		PrimaryKey: []*schema.Column{UsersTeamsColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "users_teams_users_users",
				Columns:    []*schema.Column{UsersTeamsColumns[2]},
				RefColumns: []*schema.Column{UsersColumns[0]},
				OnDelete:   schema.Cascade,
			},
			{
				Symbol:     "users_teams_teams_teams",
				Columns:    []*schema.Column{UsersTeamsColumns[3]},
				RefColumns: []*schema.Column{TeamsColumns[0]},
				OnDelete:   schema.Cascade,
			},
		},
		Indexes: []*schema.Index{
			{
				Name:    "usersteams_team_id_user_id",
				Unique:  true,
				Columns: []*schema.Column{UsersTeamsColumns[3], UsersTeamsColumns[2]},
			},
		},
	}
	// Tables holds all the tables in the schema.
	Tables = []*schema.Table{
		AccessTokensTable,
		EnvsTable,
		EnvAliasesTable,
		EnvBuildsTable,
		TeamsTable,
		TeamAPIKeysTable,
		TiersTable,
		UsersTable,
		UsersTeamsTable,
	}
)

func init() {
	AccessTokensTable.ForeignKeys[0].RefTable = UsersTable
	AccessTokensTable.Annotation = &entsql.Annotation{}
	EnvsTable.ForeignKeys[0].RefTable = TeamsTable
	EnvsTable.Annotation = &entsql.Annotation{}
	EnvAliasesTable.ForeignKeys[0].RefTable = EnvsTable
	EnvAliasesTable.Annotation = &entsql.Annotation{
		Table: "env_aliases",
	}
	EnvBuildsTable.ForeignKeys[0].RefTable = EnvsTable
	EnvBuildsTable.Annotation = &entsql.Annotation{}
	TeamsTable.ForeignKeys[0].RefTable = TiersTable
	TeamsTable.Annotation = &entsql.Annotation{}
	TeamAPIKeysTable.ForeignKeys[0].RefTable = TeamsTable
	TeamAPIKeysTable.Annotation = &entsql.Annotation{}
	TiersTable.Annotation = &entsql.Annotation{}
	TiersTable.Annotation.Checks = map[string]string{
		"tiers_concurrent_sessions_check": "concurrent_instances > 0",
		"tiers_disk_mb_check":             "disk_mb > 0",
	}
	UsersTable.Annotation = &entsql.Annotation{}
	UsersTeamsTable.ForeignKeys[0].RefTable = UsersTable
	UsersTeamsTable.ForeignKeys[1].RefTable = TeamsTable
	UsersTeamsTable.Annotation = &entsql.Annotation{}
}
