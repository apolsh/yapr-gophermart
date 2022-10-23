package postgres

import (
	"context"
	"embed"
	"os"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog/log"
	// migrate tools
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const _defaultAttempts = 5
const _defaultTimeout = 10 * time.Second

//go:embed migrations/*.sql
var fs embed.FS

func RunMigration(pool *pgxpool.Pool) {

	q := `BEGIN;
create table if not exists "user"
(
    id       uuid default gen_random_uuid() not null
    primary key,
    login    varchar(255)                   not null,
    password varchar(255)                   not null
    );

alter table "user"
    owner to gophermartuser;

create unique index if not exists user_login_uindex
    on "user" (login);

create table if not exists "order"
(
    number      bigint                  not null
    constraint order_pk
    primary key,
    status      varchar(255)             not null,
    accrual     numeric(12, 2)           not null,
    uploaded_at timestamp with time zone not null,
                              user_id     uuid
                              constraint order_user_id_fk
                              references "user"
                              on delete cascade
                              );

alter table "order"
    owner to gophermartuser;

create table if not exists withdrawal
(
    "order"      bigint                  not null
    constraint withdrawal_order_number_fk
    references "order"
    on delete cascade,
    sum          numeric(12, 2)           not null,
    processed_at timestamp with time zone not null,
                               user_id      uuid
                               constraint withdrawal_user_id_fk
                               references "user"
                               on delete cascade
                               );

alter table withdrawal
    owner to gophermartuser;

create table if not exists balance
(
    user_id   uuid
    constraint balance_user_id_fk
    references "user"
    on delete cascade,
    current   numeric(12, 2),
    withdrawn numeric(12, 2)
    );

alter table balance
    owner to gophermartuser;

create unique index if not exists balance_user_uindex
    on balance (user_id);

create or replace function create_new_balance() returns trigger
    language plpgsql
as
$$
begin
insert into balance(user_id, current, withdrawn) values (new.user_id, 0, 0)
on conflict(user_id) do nothing;

RETURN NEW;
END;
$$;

alter function create_new_balance() owner to gophermartuser;

create trigger ins_new_balance
    after insert
    on "order"
    for each row
    execute procedure create_new_balance();

create or replace function add_accrual_to_balance() returns trigger
    language plpgsql
as
$$
begin
	if new.accrual is not null then
update balance
set current = balance.current + new.accrual
where user_id = old.user_id;
end if;
return new;
END;
$$;

alter function add_accrual_to_balance() owner to gophermartuser;

create trigger add_accrual_to_balance
    after update
    on "order"
    for each row
    execute procedure add_accrual_to_balance();

create or replace function add_withdrawn_to_balance() returns trigger
    language plpgsql
as
$$
begin
	if new.sum is not null then
update balance
set current = balance.current - new.sum,
    withdrawn = balance.withdrawn + new.sum
where user_id = new.user_id;
end if;
return new;
END;
$$;

alter function add_withdrawn_to_balance() owner to gophermartuser;

create trigger add_withdrawn_to_balance
    after insert
    on withdrawal
    for each row
    execute procedure add_withdrawn_to_balance();
COMMIT;
`

	_, err := pool.Exec(context.Background(), q)
	if err != nil {
		log.Error().Err(err).Msg(err.Error())
		os.Exit(1)
	}
}
