import org.json.JSONObject;

String body = prev.getResponseDataAsString();

if (body.trim().length() == 0 || body == null) {
    AssertionResult.setFailureMessage("not response");
	AssertionResult.setFailure(true);
} else {
    JSONObject json = new JSONObject(body); 

    try {
        if (json.getInt("_id") == 0) {
            AssertionResult.setFailureMessage("i_d is 0");
            AssertionResult.setFailure(true);
        }
    } catch (JSONException e) {
        AssertionResult.setFailureMessage("body: " + body + " error: " +  e.printStackTrace());
        AssertionResult.setFailure(true);
    }
}
