defmodule PokerBackendWeb.UserRegistrationController do
  use PokerBackendWeb, :controller

  alias PokerBackend.Accounts

  def new(conn, _params) do
    render(conn, :new,
      form: Phoenix.Component.to_form(%{}, as: "user"),
      errors: []
    )
  end

  def create(conn, %{"user" => user_params}) do
    case Accounts.register_user(user_params) do
      {:ok, user} ->
        {:ok, _} =
          Accounts.deliver_login_instructions(
            user,
            &url(~p"/users/log-in/#{&1}")
          )

        conn
        |> put_flash(
          :info,
          "An email was sent to #{user.email}, please access it to confirm your account."
        )
        |> redirect(to: ~p"/users/log-in")

      {:error, %Ecto.Changeset{} = changeset} ->
        render(conn, :new,
          form:
            Phoenix.Component.to_form(
              Map.take(user_params, ["email", "username"]),
              as: "user"
            ),
          errors: changeset.errors
        )
    end
  end
end
