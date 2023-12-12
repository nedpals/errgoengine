public class OperatorCannotBeApplied {
    public static void main(String[] args) {
        String text = "Hello";
        int number = 5;

        // Operator '<' cannot be applied to String and int
        if (text < number) {}
    }
}
