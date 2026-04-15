defmodule PokerBackendWeb.TableChannel do
  use PokerBackendWeb, :channel

  @max_table_id_length 64
  @table_id_pattern ~r/^[a-zA-Z0-9_\-]+$/

  @impl true
  def join("table:" <> table_id, params, socket) do
    if valid_table_id?(table_id) do
      {:ok, _pid} = PokerBackend.Table.ensure_started(table_id)
      Phoenix.PubSub.subscribe(PokerBackend.PubSub, PokerBackend.Table.topic(table_id))

      player_name = Map.get(params, "player_name", "guest")
      player_id = Map.get(params, "player_id", "guest")
      {:ok, _state} = PokerBackend.Table.join(table_id, player_id, player_name)

      socket =
        socket
        |> assign(:table_id, table_id)
        |> assign(:player_id, player_id)

      state = PokerBackend.Table.state(table_id)
      filtered_state = filter_state_for_player(state, player_id)

      {:ok, %{type: "table_state", state: filtered_state}, socket}
    else
      {:error, %{reason: "invalid_table_id"}}
    end
  end

  @impl true
  def handle_in("ping", payload, socket) do
    {:ok, event} = PokerBackend.Table.ping(socket.assigns.table_id, payload)
    {:reply, {:ok, event}, socket}
  end

  def handle_in(_event, _payload, socket) do
    {:reply, {:error, %{error: "unsupported_event"}}, socket}
  end

  @impl true
  def handle_info({:table_event, %{type: "table_state", state: state} = payload}, socket) do
    filtered_state = filter_state_for_player(state, socket.assigns.player_id)
    push(socket, "table_event", %{payload | state: filtered_state})
    {:noreply, socket}
  end

  @impl true
  def handle_info({:table_event, payload}, socket) do
    push(socket, "table_event", payload)
    {:noreply, socket}
  end

  @impl true
  def terminate(_reason, socket) do
    _ = PokerBackend.Table.leave(socket.assigns.table_id, socket.assigns.player_id)
    :ok
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

  defp valid_table_id?(table_id) do
    byte_size(table_id) > 0 and
      byte_size(table_id) <= @max_table_id_length and
      Regex.match?(@table_id_pattern, table_id)
  end
end
