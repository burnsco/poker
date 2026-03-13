defmodule PokerBackendWeb.UserSessionJSONControllerTest do
  use PokerBackendWeb.ConnCase, async: true

  import PokerBackend.AccountsFixtures

  describe "POST /api/users/register" do
    test "registers a user and creates a session", %{conn: conn} do
      email = unique_user_email()

      conn =
        post(conn, ~p"/api/users/register", %{
          "user" => %{
            "email" => email,
            "username" => "river_runner",
            "password" => valid_user_password()
          }
        })

      assert %{
               "data" => %{
                 "email" => ^email,
                 "username" => "river_runner"
               }
             } = json_response(conn, 200)

      assert get_session(conn, :user_token)

      me_conn =
        conn
        |> recycle()
        |> get(~p"/api/users/me")

      assert %{"data" => %{"email" => ^email}} = json_response(me_conn, 200)
    end

    test "returns normalized field errors for invalid registrations", %{conn: conn} do
      conn = post(conn, ~p"/api/users/register", %{"user" => %{"email" => "bad"}})

      assert %{
               "error" => "Please fix the highlighted fields.",
               "errors" => errors
             } = json_response(conn, 422)

      assert errors["email"] == ["must have the @ sign and no spaces"]
      assert errors["username"] == ["can't be blank"]
      assert errors["password"] == ["can't be blank"]
    end
  end

  describe "POST /api/users/log-in" do
    test "logs the user in and returns the user payload", %{conn: conn} do
      user = user_fixture(%{username: "turn_shaper"}) |> set_password()

      conn =
        post(conn, ~p"/api/users/log-in", %{
          "user" => %{
            "email" => user.email,
            "password" => valid_user_password()
          }
        })

      assert %{
               "data" => %{
                 "email" => user_email,
                 "username" => "turn_shaper"
               }
             } = json_response(conn, 200)

      assert user_email == user.email
      assert get_session(conn, :user_token)
    end

    test "returns a stable error payload for invalid credentials", %{conn: conn} do
      user = user_fixture() |> set_password()

      conn =
        post(conn, ~p"/api/users/log-in", %{
          "user" => %{
            "email" => user.email,
            "password" => "invalid"
          }
        })

      assert %{
               "code" => "invalid_credentials",
               "error" => "Invalid email or password"
             } = json_response(conn, 401)
    end
  end

  describe "GET /api/users/me" do
    test "returns unauthorized when the user is not logged in", %{conn: conn} do
      conn = get(conn, ~p"/api/users/me")

      assert %{
               "code" => "not_authenticated",
               "error" => "Not authenticated"
             } = json_response(conn, 401)
    end
  end

  describe "DELETE /api/users/log-out" do
    test "clears the session and returns success", %{conn: conn} do
      user = user_fixture()

      conn =
        conn
        |> log_in_user(user)
        |> delete(~p"/api/users/log-out")

      assert %{"ok" => true} = json_response(conn, 200)
      refute get_session(conn, :user_token)
    end
  end
end
