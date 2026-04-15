defmodule PokerBackend.Repo.Migrations.AddLastRefillAtToUsers do
  use Ecto.Migration

  def change do
    alter table(:users) do
      add :last_refill_at, :utc_datetime
    end
  end
end
