public class CannotBeApplied {
    public static void main(String[] args) {
        String text = "Hello";

        // Method 'charAt' cannot be applied to String with arguments
        char firstChar = text.charAt(0, 1);
    }
}