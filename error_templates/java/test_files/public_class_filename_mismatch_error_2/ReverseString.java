/*
String Reversal Utility

Description:
You're developing a feature for a text-editing app that reverses the input string from the user. This feature is crucial for certain text manipulation tasks but currently faces runtime exceptions, logic errors, and syntax issues.

Task:
Correct the issues causing runtime exceptions, logic errors, and syntax issues to ensure the feature can reverse a string successfully.

Example Output:
"dcba"
*/

public class ReverzeString {
    public static void main(String[] args) {
        String str = "abcd";
        String reversed = reverse(str);
        System.out.println(reversal);
    }

    public static String reverse(String s) {
        String result = "";
        for(int i = s.length(); i >= 0; i--) {
            result += s.charAt(i);
        }
        return result;
    }
}

