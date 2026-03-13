defmodule PokerBackendWeb.Plugs.CORS do
  import Plug.Conn

  def init(opts), do: opts

  def call(conn, _opts) do
    origin = resolve_origin(conn)

    conn
    |> put_resp_header("access-control-allow-origin", origin)
    |> put_resp_header("access-control-allow-credentials", "true")
    |> put_resp_header("access-control-allow-methods", "GET,POST,PUT,DELETE,OPTIONS")
    |> put_resp_header(
      "access-control-allow-headers",
      "content-type,authorization,x-requested-with"
    )
    |> put_resp_header("vary", "origin")
  end

  defp resolve_origin(conn) do
    configured_origin = Application.get_env(:poker_backend, :cors_origin, "*")
    request_origin = get_req_header(conn, "origin") |> List.first()

    case configured_origin do
      "*" when is_binary(request_origin) and request_origin != "" -> request_origin
      "*" -> "*"
      origin -> origin
    end
  end
end
