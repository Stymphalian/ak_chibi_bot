import { redirect } from "react-router-dom";
import { deleteContact } from "../contacts";

export async function action(data:any) {
  const params = data.params;
  throw new Error("oh dang!");
  await deleteContact(params.contactId);
  return redirect("/");
}