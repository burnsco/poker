defmodule PokerBackendWeb.UserSessionJSONController do
  use PokerBackendWeb, :controller

  alias PokerBackend.Accounts
  alias PokerBackendWeb.UserAuth

  def create(conn, %{"user" => %{"email" => email, "password" => password} = params}) do
    if user = Accounts.get_user_by_email_and_password(email, password) do
      conn
      |> UserAuth.log_in_user_json(user, params)
    else
      conn
      |> put_status(:unauthorized)
      |> json(%{error: "Invalid email or password", code: "invalid_credentials"})
    end
  end

  def delete(conn, _params) do
    conn
    |> UserAuth.log_out_user_json()
  end

  def show(conn, _params) do
    if user = conn.assigns.current_scope && conn.assigns.current_scope.user do
      conn
      |> put_view(PokerBackendWeb.UserJSON)
      |> render(:show, user: user)
    else
      conn
      |> put_status(:unauthorized)
      |> json(%{error: "Not authenticated", code: "not_authenticated"})
    end
  end
end
