defmodule PokerBackendWeb.Plugs.CORS do
  import Plug.Conn

  def init(opts), do: opts

  def call(conn, _opts) do
    origin = Application.get_env(:poker_backend, :cors_origin, "*")

    conn
    |> put_resp_header("access-control-allow-origin", origin)
    |> put_resp_header("access-control-allow-methods", "GET,POST,OPTIONS")
    |> put_resp_header("access-control-allow-headers", "content-type,authorization")
  end
end
