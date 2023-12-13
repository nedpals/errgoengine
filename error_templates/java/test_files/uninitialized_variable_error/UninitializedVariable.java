public class UninitializedVariable {
    public static void main(String[] args) {
        int number;

        // Variable 'number' might not have been initialized
        System.out.println(number);
    }
}
