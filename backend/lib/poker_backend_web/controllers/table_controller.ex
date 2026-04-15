defmodule PokerBackendWeb.TableController do
  use PokerBackendWeb, :controller

  @max_table_id_length 64
  @table_id_pattern ~r/^[a-zA-Z0-9_\-]+$/

  plug PokerBackendWeb.Plugs.TableActionRateLimit when action in [:update_action]

  def index(conn, _params) do
    table_ids = PokerBackend.Table.list_active_tables()
    json(conn, %{data: table_ids})
  end

  def summary(conn, _params) do
    table_ids = PokerBackend.Table.list_active_tables()

    data =
      Enum.map(table_ids, fn table_id ->
        state =
          table_id
          |> PokerBackend.Table.state()
          |> filter_state_for_player(nil)

        %{table_id: table_id, state: state}
      end)

    json(conn, %{data: data})
  end

  def show(conn, %{"table_id" => table_id}) do
    if valid_table_id?(table_id) do
      {:ok, _pid} = PokerBackend.Table.ensure_started(table_id)
      viewer_player_id = conn.assigns[:current_scope] && conn.assigns.current_scope.user && conn.assigns.current_scope.user.id
      state = PokerBackend.Table.state(table_id)
      json(conn, filter_state_for_player(state, viewer_player_id))
    else
      conn |> put_status(400) |> json(%{error: "invalid_table_id"})
    end
  end

  def update_action(conn, %{"table_id" => table_id, "action" => action} = params) do
    cond do
      not valid_table_id?(table_id) ->
        conn |> put_status(400) |> json(%{error: "Invalid table id", code: "invalid_table_id"})

      is_nil(conn.assigns[:current_scope] && conn.assigns.current_scope.user) ->
        conn
        |> put_status(:unauthorized)
        |> json(%{
          error: "Log in to take a seat or manage table actions.",
          code: "authentication_required"
        })

      true ->
        user = conn.assigns.current_scope.user
        action_params = authorize_action_params(params, user)

        {:ok, _pid} = PokerBackend.Table.ensure_started(table_id)

        case PokerBackend.Table.action(table_id, action, action_params) do
          {:ok, state} -> json(conn, state)

          {:error, reason} ->
            conn
            |> put_status(422)
            |> json(%{error: format_action_error(reason), code: reason})
        end
    end
  end

  defp valid_table_id?(table_id) do
    byte_size(table_id) > 0 and
      byte_size(table_id) <= @max_table_id_length and
      Regex.match?(@table_id_pattern, table_id)
  end

  defp authorize_action_params(params, user) do
    params
    |> Map.put("player_id", user.id)
    |> Map.put("player_name", user.username)
  end

  defp format_action_error(reason) when is_binary(reason) do
    reason
    |> String.replace("_", " ")
    |> String.capitalize()
  end

  defp filter_state_for_player(state, viewer_player_id) do
    filtered_players =
      Enum.map(state.players, fn player ->
        if should_reveal_cards?(player, viewer_player_id) do
          player
        else
          %{player | hole_cards: [nil, nil]}
        end
      end)

    state
    |> Map.put(:players, filtered_players)
    |> update_in([:hand_state], fn hand_state -> Map.delete(hand_state, :deck) end)
  end

  defp should_reveal_cards?(player, viewer_player_id) do
    player.is_bot or player.show_cards or to_string(player.player_id) == to_string(viewer_player_id)
  end
end
